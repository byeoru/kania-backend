package util

import (
	"math"
	"time"

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

func Map[T any, R any](input []T, mapper func(T) R) []R {
	result := make([]R, len(input))
	for i, v := range input {
		result[i] = mapper(v)
	}
	return result
}

func Find[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true // 인덱스와 함께 true 반환
		}
	}
	return false // 찾지 못했을 경우
}

func CalculateLevyAdvanceSpeed(levy *db.Levy) float64 {
	unitStat := GetUnitStat()
	unitCount := levy.Swordmen + levy.Archers + levy.ShieldBearers + levy.Lancers + levy.SupplyTroop
	if unitCount == 0 {
		return 0
	}
	speed := float64(levy.Swordmen*unitStat.Swordman.Speed+
		levy.Archers*unitStat.Archer.Speed+
		levy.ShieldBearers*unitStat.ShieldBearer.Speed+
		levy.Lancers*unitStat.Lancer.Speed) / float64(unitCount)
	return speed
}

func IsAnnihilated(levy *db.Levy) bool {
	wholeCount := levy.Swordmen + levy.Archers + levy.ShieldBearers + levy.Lancers + levy.SupplyTroop
	return wholeCount == 0
}

func CalculateCurrentWorldTime(standardRealTime time.Time, standardWorldTime time.Time) time.Time {
	// 현재 시간
	currentRealTime := time.Now()
	// 현실 기준 경과 시간 계산
	realElapsed := currentRealTime.Sub(standardRealTime)
	// 경과 시간에 배속 배수 곱하기
	acceleratedDuration := realElapsed * time.Duration(SpeedMultiplier)
	// 현재 세계관 시간 계산
	return standardWorldTime.Add(acceleratedDuration)
}

func CalculateCompletionDate(currentWorldTime time.Time, delta float64) time.Time {
	hours := int(delta)
	minutes := math.Round((delta - float64(hours)) * 60)
	return currentWorldTime.Add(time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes))
}
