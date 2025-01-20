package cron

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/util"
)

var (
	LevyActionCronInit     sync.Once
	LevyActionCronInstance *LevyActionCron
)

type LevyActionCron struct {
	levyActionService      *service.LevyActionService
	worldTimeRecordService *service.WorldTimeRecordService
}

func NewLevyActionCron(service *service.Service) *LevyActionCron {
	LevyActionCronInit.Do(func() {
		LevyActionCronInstance = &LevyActionCron{
			levyActionService:      service.LevyActionService,
			worldTimeRecordService: service.WorldTimeRecordService,
		}
	})
	return LevyActionCronInstance
}

func (c *LevyActionCron) ExecuteCronLevyActions(ctx *context.Context, ticker *time.Ticker) {
	recordedTime, err := c.worldTimeRecordService.FindLatestWorldTime(ctx)
	var standardWorldTime time.Time
	var standardRealTime time.Time
	if err != nil {
		if err == sql.ErrNoRows {
			standardWorldTime = time.Date(312, 5, 2, 1, 20, 0, 0, time.UTC)
			standardRealTime = time.Now()
		}
	} else {
		standardWorldTime = recordedTime.WorldStoppedAt
		standardRealTime = recordedTime.CreatedAt
	}
	for range ticker.C {
		util.WorldTime = calculateCurrentWorldTime(standardRealTime, standardWorldTime)
		fmt.Println(util.WorldTime)
		err := c.levyActionService.ExecuteCronLevyActions(ctx, util.WorldTime)
		if err != nil {
			arg := db.CreateWorldTimeRecordParams{
				StopReason:     err.Error(),
				WorldStoppedAt: util.WorldTime,
			}
			c.worldTimeRecordService.CreateWorldTimeRecord(ctx, &arg)
		}
	}
}

func (c *LevyActionCron) RecordWorldTime(ctx *context.Context, currentWorldTime time.Time) {
	arg := db.CreateWorldTimeRecordParams{
		StopReason:     "stop server",
		WorldStoppedAt: currentWorldTime,
	}
	c.worldTimeRecordService.CreateWorldTimeRecord(ctx, &arg)
}

func calculateCurrentWorldTime(startRealTime time.Time, startWorldTime time.Time) time.Time {
	// 현재 시간
	currentRealTime := time.Now()
	// 현실 기준 경과 시간 계산
	realElapsed := currentRealTime.Sub(startRealTime)
	// 경과 시간에 배속 배수 곱하기
	acceleratedDuration := realElapsed * time.Duration(util.SpeedMultiplier)
	// 현재 세계관 시간 계산
	return startWorldTime.Add(acceleratedDuration)
}
