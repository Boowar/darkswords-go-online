package types

import (
	"time"
)

// --- Типы ---
type LocationType string
type BiomeType string
type MagicSchool string
type Race string
type Religion string

const (
	LocationShop  LocationType = "shop"
	LocationGuild LocationType = "guild"
	LocationStair LocationType = "stair"
	LocationEmpty LocationType = "empty"
)

const (
	BiomeRoad      BiomeType = "road"
	BiomeForest    BiomeType = "forest"
	BiomeDesert    BiomeType = "desert"
	BiomeMountains BiomeType = "mountains"
	BiomeSwamp     BiomeType = "swamp"
)

const (
	MagicFire      MagicSchool = "fire"
	MagicIce       MagicSchool = "ice"
	MagicLightning MagicSchool = "lightning"
)

const (
	RaceHuman Race = "human"
	RaceElf   Race = "elf"
	RaceDrow  Race = "drow"
	RaceOrc   Race = "orc"
	RaceDwarf Race = "dwarf"
)

const (
	ReligionOrder Religion = "order"
	ReligionChaos Religion = "chaos"
	ReligionLight Religion = "light"
	ReligionDark  Religion = "dark"
)

// --- Игровые структуры ---
type Position struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

type Player struct {
	Name          string          `json:"Name"`
	Level         int             `json:"Level"`
	CurrentXP     int             `json:"CurrentXP"`
	RequiredXP    int             `json:"RequiredXP"`
	HP            int             `json:"HP"`
	MP            int             `json:"MP"`
	Race          Race            `json:"Race"`
	Religion      Religion        `json:"Religion"`
	Position      Position        `json:"Position"`
	Body          int             `json:"Body"`
	Strength      int             `json:"Strength"`
	Dexterity     int             `json:"Dexterity"`
	Intelligence  int             `json:"Intelligence"`
	MagicSchools  []MagicSchool   `json:"MagicSchools"`
	Items         []Item          `json:"Items"`
	EquippedItems map[string]Item `json:"EquippedItems"` // weapon, ring
	IsAlive       bool            `json:"IsAlive"`
	Damage        int             `json:"Damage"`
}

type Item struct {
	Name      string      `json:"Name"`
	Type      string      `json:"Type"`
	MinLevel  int         `json:"MinLevel"`
	Bonus     interface{} `json:"Bonus"`
	Charge    int         `json:"Charge"`
	MaxCharge int         `json:"MaxCharge"`
}

type Monster struct {
	Name   string              `json:"Name"`
	Level  int                 `json:"Level"`
	HP     int                 `json:"HP"`
	MaxHP  int                 `json:"MaxHP"`
	Damage int                 `json:"Damage"`
	School MagicSchool         `json:"School"`
	Resist map[MagicSchool]int `json:"Resist"`
	IsBoss bool                `json:"IsBoss"`
}

type Room struct {
	X                int          `json:"X"`
	Y                int          `json:"Y"`
	BiomeType        BiomeType    `json:"BiomeType"`    // road, forest, desert, mountains, swamp
	LocationType     LocationType `json:"LocationType"` // shop, guild, stair, empty
	Monsters         []Monster    `json:"Monsters"`
	NPCs             []string     `json:"NPCs"`
	CorpsePlayer     string       `json:"CorpsePlayer"`
	CorpseTime       time.Time    `json:"CorpseTime"`
	Items            []Item       `json:"Items"`        // предметы в комнате
	PlayersItems     []Item       `json:"PlayersItems"` // из трупа
	LastMonsterSpawn time.Time    `json:"LastMonsterSpawn"`
}

type GameState struct {
	Players []Player `json:"Players"`
	Map     [][]Room `json:"Map"`
	Log     []string `json:"Log"`
}

type Battle struct {
	RoomX, RoomY int
	BattleID     string
	Participants map[string]bool
	TurnTimer    *time.Timer
}
