package game

import (
	"dark-swords/types"
	"dark-swords/utils"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func calculatePlayerDamage(p types.Player) int {
	damage := p.Strength
	for _, item := range p.EquippedItems {
		switch v := item.Bonus.(type) {
		case map[string]interface{}:
			if dmg, ok := v["damage"].(float64); ok {
				damage += int(dmg)
			}
		case map[string]int:
			damage += v["damage"]
		case int:
			damage += v
		}
	}
	return damage
}

func findPlayerIndex(name string) int {
	for i, p := range GameState.Players {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func HandleMove(conn *websocket.Conn, data json.RawMessage) {
	type MoveData struct {
		Name string `json:"name"`
		To   struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"to"`
	}

	var payload MoveData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных MOVE: %v", err)
		return
	}

	playerName := payload.Name
	to := payload.To

	Mutex.Lock()
	Mutex.Unlock()

	for i, p := range GameState.Players {
		if p.Name == playerName {
			oldX, oldY := p.Position.X, p.Position.Y
			newX, newY := to.X, to.Y

			// Проверяем границы
			if newX < 0 || newX >= 10 || newY < 0 || newY >= 10 {
				log.Printf("🚫 Попытка выхода за пределы карты")
				return
			}

			//Проверяем безопасную зону
			toType := GameState.Map[newY][newX].BiomeType

			if toType == "empty" {
				log.Printf("🚫 Попытка выхода за пределы карты")
				return
			}

			// Обновляем позицию
			GameState.Map[oldY][oldX].CorpsePlayer = ""
			GameState.Map[newY][newX].CorpsePlayer = ""

			GameState.Players[i].Position.X = newX
			GameState.Players[i].Position.Y = newY
			GameState.Players[i].CurrentXP += 1 // тестовый опыт за ход

			checkLevelUp(i)

			break
		}
	}

	BroadcastGameState()
}

func checkLevelUp(index int) {
	player := &GameState.Players[index]
	if player.CurrentXP >= player.RequiredXP {
		player.Level++
		player.RequiredXP = player.Level * 10
		player.Body += 1
		player.Strength += 1
		player.Dexterity += 1
		player.Intelligence += 1
		player.MP += player.Intelligence * 2
		player.HP = player.Body*3 + player.Strength*2
		player.Damage = calculatePhysicalDamage(*player)
		log.Printf("🎉 %s повысил уровень до %d!", player.Name, player.Level)
	}
}

func calculatePhysicalDamage(p types.Player) int {
	return p.Strength + sumItemPower(p.EquippedItems)
}
func sumItemPower(items map[string]types.Item) int {
	total := 0
	for _, item := range items {
		switch v := item.Bonus.(type) {
		case map[string]interface{}:
			if dmg, ok := v["damage"].(float64); ok {
				total += int(dmg)
			}
		case map[string]int:
			total += v["damage"]
		case int:
			total += v
		}
	}
	return total
}

func CheckCorpseLifespan() {
	for {
		time.Sleep(time.Minute) // раз в минуту обновляем карту
		Mutex.Lock()
		for y := range GameState.Map {
			for x := range GameState.Map[y] {
				if GameState.Map[y][x].CorpseTime.Before(time.Now()) && GameState.Map[y][x].CorpsePlayer != "" {
					log.Printf("☠️ Труп %s исчез из [%d,%d]", GameState.Map[y][x].CorpsePlayer, x, y)
					GameState.Map[y][x].Items = append(GameState.Map[y][x].Items, GameState.Map[y][x].PlayersItems...)
					GameState.Map[y][x].CorpsePlayer = ""
					GameState.Map[y][x].PlayersItems = nil
				}
			}
		}
		Mutex.Unlock()
	}
}

func HandleJoin(conn *websocket.Conn, data json.RawMessage) {
	type JoinData struct {
		Name     string         `json:"name"`
		Race     types.Race     `json:"race"`
		Religion types.Religion `json:"religion"`
	}

	var payload JoinData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных JOIN: %v", err)
		return
	}

	playerName := payload.Name
	race := payload.Race
	religion := payload.Religion
	Mutex.Lock()
	log.Printf("[handleJoin] 🔒 Мьютекс захвачен")
	Mutex.Unlock()
	log.Printf("[handleJoin] 🔓 Мьютекс освобождён")

	// Проверяем, существует ли уже такой игрок
	for _, p := range GameState.Players {
		if p.Name == playerName {
			log.Printf("⚠️ Игрок '%s' уже существует", playerName)
			return
		}
	}

	// Базовые характеристики по расе
	body, strength, dexterity, intelligence := 10, 10, 10, 10
	switch race {
	case types.RaceHuman:
		strength += 1
		dexterity += 1
	case types.RaceElf:
		intelligence += 2
		dexterity += 1
	case types.RaceDrow:
		dexterity += 2
		body -= 1
	case types.RaceOrc:
		body += 2
		intelligence -= 2
	case types.RaceDwarf:
		body += 3
		strength += 1
	}

	newPlayer := types.Player{
		Name:          playerName,
		Level:         1,
		CurrentXP:     0,
		RequiredXP:    10,
		HP:            body*3 + strength*2,
		MP:            intelligence * 5,
		Race:          race,
		Religion:      religion,
		Position:      types.Position{X: 0, Y: 0},
		Body:          body,
		Strength:      strength,
		Dexterity:     dexterity,
		Intelligence:  intelligence,
		MagicSchools:  []types.MagicSchool{},
		Items:         []types.Item{},
		EquippedItems: make(map[string]types.Item),
		IsAlive:       true,
	}
	newPlayer.Damage = calculatePhysicalDamage(newPlayer)

	GameState.Players = append(GameState.Players, newPlayer)
	Clients[conn] = playerName

	log.Printf("✅ Игрок '%s' присоединился к миру как %s (%s)", playerName, race, religion)
	BroadcastGameState()
}

func HandleLootCorpse(conn *websocket.Conn, data json.RawMessage) {
	type LootData struct {
		Name string `json:"name"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}
	var payload LootData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных LOOT: %v", err)
		return
	}

	playerName := payload.Name
	x, y := payload.X, payload.Y

	Mutex.Lock()
	Mutex.Unlock()

	playerIndex := -1
	for i, p := range GameState.Players {
		if p.Name == playerName {
			playerIndex = i
			break
		}
	}
	if playerIndex == -1 {
		return
	}

	// Проверяем, есть ли труп в этой комнате
	corpse := &GameState.Map[y][x]
	if corpse.CorpsePlayer == "" || !corpse.CorpseTime.After(time.Now()) {
		log.Printf("🚫 Нет трупа в [%d,%d]", x, y)
		return
	}

	// Игрок должен быть в той же комнате
	if GameState.Players[playerIndex].Position.X != x || GameState.Players[playerIndex].Position.Y != y {
		log.Printf("🚫 Игрок %s не в комнате [%d,%d]", playerName, x, y)
		return
	}

	// Забираем предметы из трупа
	player := &GameState.Players[playerIndex]

	// Предметы трупа
	corpseItems := GameState.Map[y][x].PlayersItems

	// Добавляем их игроку
	player.Items = append(player.Items, corpseItems...)

	// Логируем событие
	log.Printf("🧟‍♂️ %s обыскал труп %s в [%d,%d]", playerName, corpse.CorpsePlayer, x, y)

	// Отправляем сообщение в чат
	utils.SendChat(fmt.Sprintf("%s обыскал труп %s", playerName, corpse.CorpsePlayer))

	// Очищаем труп
	GameState.Map[y][x].PlayersItems = nil
	GameState.Map[y][x].CorpsePlayer = ""
	GameState.Map[y][x].CorpseTime = time.Time{}

	BroadcastGameState()
}

func HandleEquip(conn *websocket.Conn, data json.RawMessage) {
	type EquipData struct {
		Name     string `json:"name"`
		ItemName string `json:"item_name"`
	}

	var payload EquipData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных EQUIP: %v", err)
		return
	}

	playerName := payload.Name
	itemName := payload.ItemName

	Mutex.Lock()
	Mutex.Unlock()

	playerIndex := findPlayerIndex(playerName)
	if playerIndex == -1 {
		log.Printf("🚫 Игрок %s не найден", playerName)
		return
	}

	player := &GameState.Players[playerIndex]

	// Убедимся, что карта EquippedItems инициализирована
	if player.EquippedItems == nil {
		player.EquippedItems = make(map[string]types.Item)
	}

	// Поиск предмета в инвентаре
	for i, item := range player.Items {
		if item.Name == itemName {
			// Проверяем тип
			if item.Type == "weapon" || item.Type == "ring" {
				// Проверяем, нет ли уже надетого такого типа
				if equipped, ok := player.EquippedItems[item.Type]; ok {
					log.Printf("⚠️ У вас уже надето: %s", equipped.Name)
					break
				}

				// Надеваем
				player.EquippedItems[item.Type] = item
				// Убираем из инвентаря
				player.Items = append(player.Items[:i], player.Items[i+1:]...)
				player.Damage = calculatePhysicalDamage(*player)

				log.Printf("🧦 %s надел %s", playerName, itemName)
				utils.SendChat(fmt.Sprintf("%s надел %s", playerName, itemName))
			} else if item.Type == "potion" || item.Type == "scroll" {
				// 💥 Используем зелье/свиток
				useItem(player, item)
				// Удаляем из инвентаря
				player.Items = append(player.Items[:i], player.Items[i+1:]...)
				log.Printf("🧪 %s использовал %s", playerName, itemName)
				utils.SendChat(fmt.Sprintf("%s использовал %s", playerName, itemName))
			} else {
				log.Printf("🚫 Неверный тип предмета: %s", item.Type)
			}

			break
		}
	}

	BroadcastGameState()
}

func HandleUnequip(conn *websocket.Conn, data json.RawMessage) {
	type UnequipData struct {
		Name     string `json:"name"`
		ItemType string `json:"item_type"` // "weapon", "ring"
	}

	var payload UnequipData
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Printf("❌ Ошибка парсинга данных UNEQUIP: %v", err)
		return
	}

	playerName := payload.Name
	itemType := payload.ItemType

	Mutex.Lock()
	Mutex.Unlock()

	playerIndex := findPlayerIndex(playerName)
	if playerIndex == -1 {
		log.Printf("🚫 Игрок %s не найден", playerName)
		return
	}

	player := &GameState.Players[playerIndex]

	// Проверяем, есть ли надетый предмет
	equippedItem, exists := player.EquippedItems[itemType]
	if !exists {
		log.Printf("🚫 Нет надетого предмета типа %s у %s", itemType, playerName)
		return
	}

	// Снимаем предмет
	delete(player.EquippedItems, itemType)
	player.Items = append(player.Items, equippedItem)
	player.Damage = calculatePhysicalDamage(*player)

	log.Printf("🧦 %s снял %s", playerName, equippedItem.Name)
	utils.SendChat(fmt.Sprintf("%s снял %s", playerName, equippedItem.Name))

	BroadcastGameState()
}

func useItem(p *types.Player, item types.Item) {
	switch item.Type {
	case "potion":
		p.HP += item.Bonus.(map[string]int)["heal"]
		if p.HP > p.Body*3+p.Strength*2 {
			p.HP = p.Body*3 + p.Strength*2
		}
	case "scroll":
		// Например, телепортация
		x, y := p.Position.X, p.Position.Y
		if x < 9 {
			p.Position.X++
		} else if y < 9 {
			p.Position.Y++
		} else {
			p.Position.X = 0
			p.Position.Y = 0
		}
	}
}

func RemoveClient(conn *websocket.Conn) {
	Mutex.Lock()
	Mutex.Unlock()

	name := Clients[conn]
	delete(Clients, conn)

	// Удаляем игрока из списка
	var updated []types.Player
	for _, p := range GameState.Players {
		if p.Name != name {
			updated = append(updated, p)
		} else {
			// Делаем его "трупом"
			p.IsAlive = false
			p.HP = 0
			x, y := p.Position.X, p.Position.Y
			GameState.Map[y][x].CorpsePlayer = p.Name
			GameState.Map[y][x].CorpseTime = time.Now().Add(24 * time.Hour)
			GameState.Map[y][x].PlayersItems = p.Items
		}
	}
	GameState.Players = updated

	log.Printf("🗑️ Клиент '%s' отключен", name)
	BroadcastGameState()
}
