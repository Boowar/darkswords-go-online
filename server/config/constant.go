package config

import (
	"dark-swords/types"
)

// Школы магии
type MagicSchool string

const (
	MagicFire      MagicSchool = "fire"
	MagicIce       MagicSchool = "ice"
	MagicLightning MagicSchool = "lightning"
)

// Локации
type LocationType string

const (
	LocationShop  LocationType = "shop"
	LocationGuild LocationType = "guild"
	LocationStair LocationType = "stair"
)

// Биомы
type BiomeType string

const (
	BiomeRoad      BiomeType = "road"
	BiomeForest    BiomeType = "forest"
	BiomeDesert    BiomeType = "desert"
	BiomeMountains BiomeType = "mountains"
	BiomeSwamp     BiomeType = "swamp"
)

// Расы
type Race string

const (
	RaceHuman Race = "human"
	RaceElf   Race = "elf"
	RaceDrow  Race = "drow"
	RaceOrc   Race = "orc"
	RaceDwarf Race = "dwarf"
)

// Религии
type Religion string

const (
	ReligionOrder Religion = "order"
	ReligionChaos Religion = "chaos"
	ReligionLight Religion = "light"
	ReligionDark  Religion = "dark"
)

var MonsterLootTable = []types.Item{
	{
		Name:     "Зелье здоровья",
		Type:     "potion",
		MinLevel: 1,
		Bonus:    map[string]int{"heal": 20},
	},
	{
		Name:     "Железный меч",
		Type:     "weapon",
		MinLevel: 3,
		Bonus:    map[string]int{"damage": 5},
	},
	{
		Name:     "Магический посох",
		Type:     "weapon",
		MinLevel: 5,
		Bonus:    map[string]int{"damage": 7},
	},
	{
		Name:     "Кольцо силы",
		Type:     "ring",
		MinLevel: 4,
		Bonus:    map[string]int{"damage": 2},
	},
}
