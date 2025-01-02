package util

const (
	TribePopulationGrowthRate   = 0.02
	KingdomPopulationGrowthRate = 0.03
	DefaultStateCoffers         = 5000
	DefaultPrivateMoney         = 0
	DefaultMorale               = 80
)

// 행군 속도
const (
	SwordmanSpeed     = 4 // 검병
	ShieldBearerSpeed = 3 // 방패병
	ArcherSpeed       = 5 // 궁병
	LancerSpeed       = 8 // 창기병
	SupplyTroopSpeed  = 4 // 보급병
)

// 전투력
const (
	SwordmanOffensive     = 13 // 검병
	ArcherOffensive       = 10 // 궁병
	ShieldBearerOffensive = 8  // 방패병
	LancerOffensive       = 18 // 창기병
	SupplyTroopOffensive  = 2  // 보급병
)

// 방어력
const (
	SwordmanDefensive     = 6  // 검병
	ArcherDefensive       = 4  // 궁병
	ShieldBearerDefensive = 12 // 방패병
	LancerDefensive       = 9  // 창기병
	SupplyTroopDefensive  = 2  // 보급병
)

// 군사 생산 비용
const (
	SwordsmanProductionCost    = 50  // 검병
	ArcherProductionCost       = 65  // 궁병
	ShieldBearerProductionCost = 80  // 방패병
	LancerProductionCost       = 105 // 창기병
	SupplyTroopProductionCost  = 30  // 보급병
)

var PoliticalEntitySet = map[string]struct{}{
	"Tribe":                {}, // 부족
	"TribalConfederation":  {}, // 부족 연맹
	"Kingdom":              {}, // 왕국
	"KingdomConfederation": {}, // 왕국 연맹
	"Empire":               {}, // 제국
	"FeudatoryState":       {}, // 번국
}

const (
	Chief    = "Chief"    // 부족장
	Monarch  = "Monarch"  // 왕, 황제
	Courtier = "Courtier" // 신하
)
