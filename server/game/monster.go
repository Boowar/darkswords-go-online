package game

import (
	"dark-swords/types"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func countMonstersInBiome(biome types.BiomeType) int {
	count := 0
	for y := range GameState.Map {
		for x := range GameState.Map[y] {
			if GameState.Map[y][x].BiomeType == biome {
				count += len(GameState.Map[y][x].Monsters)
			}
		}
	}
	return count
}

func spawnMonstersInBiome(biome types.BiomeType) {
	Mutex.Lock()
	Mutex.Unlock()

	toSpawn := 10 - countMonstersInBiome(biome)
	if toSpawn <= 0 {
		return
	}

	var startX, endX, startY, endY int
	switch biome {
	case types.BiomeForest:
		startX, endX = 0, 3
		startY, endY = 0, 3
	case types.BiomeDesert:
		startX, endX = 5, 8
		startY, endY = 0, 3
	case types.BiomeMountains:
		startX, endX = 0, 3
		startY, endY = 5, 8
	case types.BiomeSwamp:
		startX, endX = 5, 8
		startY, endY = 5, 8
	default:
		return
	}

	spawned := 0
	for y := startY; y <= endY && spawned < toSpawn; y++ {
		for x := startX; x <= endX && spawned < toSpawn; x++ {
			room := &GameState.Map[y][x]
			if room.LocationType != "" {
				continue
			}
			if len(room.Monsters) >= 10 {
				continue
			}

			level := rand.Intn(5) + 1
			damage := level * 3
			hp := level * 5

			room.Monsters = append(room.Monsters, types.Monster{
				Name:   "Монстр",
				Level:  level,
				HP:     hp,
				MaxHP:  hp,
				Damage: damage,
				School: types.MagicFire,
				Resist: map[types.MagicSchool]int{
					types.MagicFire:      5,
					types.MagicIce:       10,
					types.MagicLightning: 2,
				},
				IsBoss: false,
			})
			spawned++
		}
	}

	log.Printf("👹 В биоме %s добавлено %d монстров", biome, spawned)
	GameState.Log = append(GameState.Log, fmt.Sprintf("👹 В биоме %s добавлено %d монстров", biome, spawned))
	BroadcastGameState()
}

func StartMonsterRespawnLoop() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			<-ticker.C
			respawnMonsters()
			BroadcastGameState()
		}
	}()
}

func respawnMonsters() {
	Mutex.Lock()
	Mutex.Unlock()

	log.Println("🔄 Проверяем биомы: возрождение монстров")

	if GameState.Map == nil {
		log.Printf("⚠️ Карта не инициализирована → пропускаем спавн монстров")
		return
	}

	biomes := []types.BiomeType{
		types.BiomeForest,
		types.BiomeDesert,
		types.BiomeMountains,
		types.BiomeSwamp,
	}

	for _, b := range biomes {
		total := countMonstersInBiome(b)
		if total < 10 {
			log.Printf("⚠️ В биоме %s всего %d монстров → нужно добавить %d", b, total, 10-total)
			spawnMonstersInBiome(b)
		}
	}

	// --- Дороги: проверяем каждую ячейку ---
	for y := range GameState.Map {
		for x := range GameState.Map[y] {
			room := &GameState.Map[y][x]
			if room.BiomeType == types.BiomeRoad && len(room.Monsters) < 2 {
				level := rand.Intn(2) + 1
				damage := level * 3
				hp := level * 5

				room.Monsters = append(room.Monsters, types.Monster{
					Name:   "Дорожный разбойник",
					Level:  level,
					HP:     hp,
					MaxHP:  hp,
					Damage: damage,
					School: types.MagicFire,
					Resist: map[types.MagicSchool]int{
						types.MagicFire:      5,
						types.MagicIce:       10,
						types.MagicLightning: 2,
					},
					IsBoss: false,
				})

				log.Printf("👹 [%d,%d] — дорожный монстр появился", x, y)
			}
		}
	}
}
