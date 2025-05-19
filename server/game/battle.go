package game

import (
	"dark-swords/config"
	"dark-swords/types"
	"dark-swords/utils"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

// --- Глобальные переменные ---
var Battles map[string]*types.Battle = make(map[string]*types.Battle)

// --- Установка данных ---
func SetGameBattle(bt map[string]*types.Battle) {
	Battles = bt
}

func HandleAttackMonster(conn *websocket.Conn, data json.RawMessage) {
	type AttackData struct {
		Name string `json:"name"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}

	var payload AttackData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных атаки: %v", err)
		return
	}

	playerName := payload.Name
	x, y := payload.X, payload.Y

	Mutex.Lock()
	Mutex.Unlock()

	// Найдём игрока
	playerIndex := -1
	for i, p := range GameState.Players {
		if p.Name == playerName && p.Position.X == x && p.Position.Y == y {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 {
		log.Printf("🚫 Игрок %s не найден или не в комнате [%d,%d]", playerName, x, y)
		return
	}

	// Проверяем, есть ли монстр в этой комнате
	if len(GameState.Map[y][x].Monsters) == 0 {
		log.Printf("🚫 В комнате [%d,%d] нет монстров", x, y)
		return
	}

	// Создаем/обновляем бой
	BattleKey := fmt.Sprintf("%d,%d", x, y)
	if _, exists := Battles[BattleKey]; !exists {
		Battles[BattleKey] = &types.Battle{
			RoomX:        x,
			RoomY:        y,
			BattleID:     BattleKey,
			Participants: map[string]bool{playerName: true},
			TurnTimer:    time.AfterFunc(2*time.Second, func() { startBattleRound(x, y) }),
		}
		log.Printf("⚔️ Бой начат в [%d,%d]", x, y)
	} else {
		Battles[BattleKey].Participants[playerName] = true
		log.Printf("🤝 %s присоединился к бою в [%d,%d]", playerName, x, y)
	}

	BroadcastGameState()
}

func startBattleRound(x, y int) {
	Mutex.Lock()
	Mutex.Unlock()

	BattleKey := fmt.Sprintf("%d,%d", x, y)
	Battle := Battles[BattleKey]
	if Battle == nil || len(Battle.Participants) == 0 {
		delete(Battles, BattleKey)
		return
	}

	// Получаем всех участников
	var attackers []types.Player
	var targetMonster types.Monster
	for _, p := range GameState.Players {
		if Battle.Participants[p.Name] {
			attackers = append(attackers, p)
		}
	}

	targetMonster = GameState.Map[y][x].Monsters[0]
	// --- Логика боя ---
	var logMessages []string

	// Бой
	totalPlayerDamage := 0

	for _, p := range attackers {
		damage := calculatePhysicalDamage(p)
		totalPlayerDamage += damage
		log.Printf("🗡️ %s наносит %d урона", p.Name, damage)
		logMessages = append(logMessages, fmt.Sprintf("🗡️ %s наносит %d урона", p.Name, damage))
	}

	// Наносим урон монстру
	targetMonster.HP -= totalPlayerDamage
	log.Printf("👹 Монстр '%s' HP: %d", targetMonster.Name, targetMonster.HP)
	logMessages = append(logMessages, fmt.Sprintf("👹 %s HP: %d", targetMonster.Name, targetMonster.HP))

	// Если монстр мёртв → награда
	if targetMonster.HP <= 0 {
		log.Printf("💀 Монстр '%s' побеждён!", targetMonster.Name)
		GameState.Map[y][x].Monsters = GameState.Map[y][x].Monsters[1:] // удаляем монстра из комнаты

		// --- Логика выпадения трофея ---
		loot := getRandomLoot()
		if loot.Name != "" {
			for name := range Battle.Participants {
				i := findPlayerIndex(name)
				if i == -1 {
					continue
				}

				GameState.Players[i].Items = append(GameState.Players[i].Items, loot)

				log.Printf("🎁 %s получил трофей: %s", name, loot.Name)
				utils.SendChat(fmt.Sprintf("%s получил трофей: %s", name, loot.Name))
			}
		}

		// Делим опыт между участниками
		xpReward := targetMonster.Level * 3
		for name := range Battle.Participants {
			for i, p := range GameState.Players {
				if p.Name == name {
					p.CurrentXP += xpReward
					checkLevelUp(i)
					GameState.Players[i] = p
					log.Printf("🏅 %s получил %d опыта", p.Name, xpReward)
					logMessages = append(logMessages, fmt.Sprintf("🏅 %s получил %d опыта", p.Name, xpReward))
				}
			}
		}

		GameState.Log = append(GameState.Log, logMessages...)
		GameState.Log = append(GameState.Log, fmt.Sprintf("💀 Монстр '%s' побеждён!", targetMonster.Name))
		delete(Battles, BattleKey)
		BroadcastGameState()
		return
	} else {
		GameState.Map[y][x].Monsters[0] = targetMonster
	}

	// Теперь монстр атакует
	monsterDamage := targetMonster.Damage
	for _, p := range attackers {
		i := findPlayerIndex(p.Name)
		GameState.Players[i].HP -= monsterDamage
		log.Printf("👹 Атакует %s → %d урона", p.Name, monsterDamage)
		logMessages = append(logMessages, fmt.Sprintf("👹 %s атакует %s. Урон: %d", targetMonster.Name, p.Name, monsterDamage))

		if GameState.Players[i].HP <= 0 {
			GameState.Players[i].IsAlive = false
			GameState.Map[GameState.Players[i].Position.Y][GameState.Players[i].Position.X].CorpsePlayer = p.Name
			GameState.Map[GameState.Players[i].Position.Y][GameState.Players[i].Position.X].CorpseTime = time.Now().Add(24 * time.Hour)
			log.Printf("☠️ %s погибает в бою", p.Name)
			logMessages = append(logMessages, fmt.Sprintf("☠️ %s погибает в бою", p.Name))
		}
	}

	// Добавляем лог боя в состояние игры
	GameState.Log = append(GameState.Log, logMessages...)

	BroadcastGameState()

	// Продолжаем бой каждые 2 секунды
	Battle.TurnTimer.Reset(2 * time.Second)
}

func getRandomLoot() types.Item {
	// Шанс выпадения трофея (например, 60%)
	if rand.Intn(100) > 60 {
		return types.Item{} // ничего не выпало
	}

	item := config.MonsterLootTable[rand.Intn(len(config.MonsterLootTable))]
	return item
}
