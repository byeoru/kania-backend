package util

import "time"

const (
	TribePopulationGrowthRate   = 0.02
	KingdomPopulationGrowthRate = 0.03
	DefaultStateCoffers         = 5000
	DefaultPrivateMoney         = 0
	DefaultMorale               = 80
)

type unit struct {
	Swordman     *UnitStat
	Archer       *UnitStat
	ShieldBearer *UnitStat
	Lancer       *UnitStat
	SupplyTroop  *UnitStat
}
type UnitStat struct {
	Speed          int32
	Offensive      int32
	Defensive      int32
	HP             int32
	ProductionCost int32
}

type ExtraStat struct {
	AddedOffensive int32
	AddedDefensive int32
}

var unitStatInstance unit = unit{
	// 검병
	Swordman: &UnitStat{
		Speed:          4,
		Offensive:      13,
		Defensive:      4,
		HP:             20,
		ProductionCost: 50,
	},
	// 궁병
	Archer: &UnitStat{
		Speed:          5,
		Offensive:      11,
		Defensive:      3,
		HP:             20,
		ProductionCost: 65,
	},
	// 방패병
	ShieldBearer: &UnitStat{
		Speed:          3,
		Offensive:      9,
		Defensive:      9,
		HP:             25,
		ProductionCost: 80,
	},
	// 창기병
	Lancer: &UnitStat{
		Speed:          8,
		Offensive:      18,
		Defensive:      7,
		HP:             30,
		ProductionCost: 105,
	},
	// 보급병
	SupplyTroop: &UnitStat{
		Speed:          4,
		Offensive:      5,
		Defensive:      2,
		HP:             20,
		ProductionCost: 30,
	},
}

func GetUnitStat() *unit {
	return &unitStatInstance
}

type AnnexResult = int8

// 합병 결과
const (
	Annihilation        AnnexResult = iota // 멸망
	Captured                               // 일반 점령
	CapitalCaptured                        // 수도 함락
	AllCapitalsCaptured                    // 모든 수도 함락
	Error                                  // 에러
)

// 부대 행동
const (
	Attack = "Attack" // 공격
	Move   = "Move"   // 주둔지 이동
	Return = "Return" // 전투 후 복귀
)

type PoliticalEntity string

const (
	Tribe                string = "Tribe"
	TribalConfederation  string = "TribalConfederation"
	Kingdom              string = "Kingdom"
	KingdomConfederation string = "KingdomConfederation"
	Empire               string = "Empire"
	FeudatoryState       string = "FeudatoryState"
	RestorationForces    string = "RestorationForces"
)

var PoliticalEntitySet = map[string]struct{}{
	"Tribe":                {}, // 부족
	"TribalConfederation":  {}, // 부족 연맹
	"Kingdom":              {}, // 왕국
	"KingdomConfederation": {}, // 왕국 연맹
	"Empire":               {}, // 제국
	"FeudatoryState":       {}, // 번국
	"RestorationForces":    {}, // 부흥세력
}

const (
	Chief    = "Chief"    // 부족장
	Monarch  = "Monarch"  // 왕, 황제
	Courtier = "Courtier" // 신하
	General  = "General"  // 장군
)

// 투항 이유
const (
	RealmAnnihilation = "RealmAnnihilation" // 국가 멸망
)

// world time
var WorldTime time.Time
var SpeedMultiplier = 60
var StandardWorldTime time.Time
var StandardRealTime time.Time
