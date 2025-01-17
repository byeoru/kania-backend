package util

import (
	db "github.com/byeoru/kania/db/repository"
)

func Min[T ~int | ~int8 | ~int16 | ~int32 | ~int64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T ~int | ~int8 | ~int16 | ~int32 | ~int64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func CalculateLevyAdvanceSpeed(levy *db.Levy) float64 {
	unitStat := GetUnitStat()
	unitCount := levy.Swordmen + levy.Archers + levy.ShieldBearers + levy.Lancers + levy.SupplyTroop
	speed := float64(levy.Swordmen*unitStat.Swordman.Speed+
		levy.Archers*unitStat.Archer.Speed+
		levy.ShieldBearers*unitStat.ShieldBearer.Speed+
		levy.Lancers*unitStat.Lancer.Speed) / float64(unitCount)
	return speed
}
