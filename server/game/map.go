package game

import (
	"dark-swords/types"
	"math/rand"
	"time"
)

func GenerateFixedMap() [][]types.Room {
	const size = 9
	mapData := make([][]types.Room, size)
	for i := 0; i < size; i++ {
		mapData[i] = make([]types.Room, size)
		for j := 0; j < size; j++ {
			mapData[i][j] = types.Room{
				X:                i,
				Y:                j,
				BiomeType:        "",
				LocationType:     "",
				Monsters:         []types.Monster{},
				Items:            []types.Item{},
				CorpseTime:       time.Time{},
				LastMonsterSpawn: time.Now().Add(-time.Hour), // чтобы сразу начали спавниться
			}
		}
	}

	// --- Биомы 4x4 ---
	biomes := []struct {
		startX, startY int
		endX, endY     int
		name           types.BiomeType
	}{
		{0, 0, 3, 3, types.BiomeForest},
		{5, 0, 8, 3, types.BiomeDesert},
		{0, 5, 3, 8, types.BiomeMountains},
		{5, 5, 8, 8, types.BiomeSwamp},
	}

	for _, b := range biomes {
		for y := b.startY; y <= b.endY; y++ {
			for x := b.startX; x <= b.endX; x++ {
				mapData[y][x].BiomeType = b.name
			}
		}
	}

	// --- Дороги между биомами ---
	roads := []struct {
		x, y int
	}{
		{4, 1}, {1, 4}, {7, 4}, {4, 7},
	}

	for _, r := range roads {
		mapData[r.y][r.x].BiomeType = types.BiomeRoad
		mapData[r.y][r.x].LocationType = ""
	}

	// --- Безопасные объекты внутри биомов ---
	objectPositions := []struct {
		biome   types.BiomeType
		x, y    int
		objType types.LocationType
	}{
		// Лес
		{types.BiomeForest, 1, 1, types.LocationShop},
		{types.BiomeForest, 3, 3, types.LocationGuild},
		{types.BiomeForest, 2, 2, types.LocationStair},

		// Пустыня
		{types.BiomeDesert, 6, 1, types.LocationShop},
		{types.BiomeDesert, 7, 2, types.LocationGuild},
		{types.BiomeDesert, 6, 2, types.LocationStair},

		// Горы
		{types.BiomeMountains, 1, 6, types.LocationShop},
		{types.BiomeMountains, 2, 7, types.LocationGuild},
		{types.BiomeMountains, 1, 7, types.LocationStair},

		// Болото
		{types.BiomeSwamp, 6, 6, types.LocationShop},
		{types.BiomeSwamp, 7, 7, types.LocationGuild},
		{types.BiomeSwamp, 6, 7, types.LocationStair},
	}

	for _, pos := range objectPositions {
		mapData[pos.y][pos.x].LocationType = pos.objType
		mapData[pos.y][pos.x].Monsters = nil
	}

	// --- Начальное наполнение монстрами ---
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			biomeType := mapData[y][x].BiomeType

			switch biomeType {
			case types.BiomeForest, types.BiomeDesert, types.BiomeMountains, types.BiomeSwamp:
				if rand.Intn(100) < 30 { // 30% шанс монстра
					level := rand.Intn(5) + 1
					damage := level * 3
					hp := level * 5
					mapData[y][x].Monsters = []types.Monster{
						{
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
						},
					}
				}
			case types.BiomeRoad:
				if rand.Intn(100) < 50 { // 20% шанс монстра на дороге
					level := rand.Intn(2) + 1
					damage := level * 3
					hp := level * 5
					mapData[y][x].Monsters = []types.Monster{
						{
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
						},
					}
				}
			}
		}
	}

	return mapData
}

func updateMapWithPlayers(players []types.Player, mapGrid [][]types.Room) [][]types.Room {
	// Очищаем все позиции игроков
	for y := range mapGrid {
		for x := range mapGrid[y] {
			mapGrid[y][x].NPCs = []string{}
		}
	}

	// Добавляем игроков в соответствующие комнаты
	for _, p := range players {
		x := p.Position.X
		y := p.Position.Y
		if x >= 0 && y >= 0 && y < len(mapGrid) && x < len(mapGrid[y]) {
			mapGrid[y][x].NPCs = append(mapGrid[y][x].NPCs, p.Name)
		}
	}

	return mapGrid
}
