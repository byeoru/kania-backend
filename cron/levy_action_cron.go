package cron

import (
	"context"
	"database/sql"
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

	if err != nil {
		if err == sql.ErrNoRows {
			util.StandardWorldTime = time.Date(312, 5, 2, 1, 20, 0, 0, time.UTC)
			util.StandardRealTime = time.Now()
		}
	} else {
		util.StandardWorldTime = recordedTime.WorldStoppedAt
		util.StandardRealTime = recordedTime.CreatedAt
	}

	for range ticker.C {
		worldTime := util.CalculateCurrentWorldTime(util.StandardRealTime, util.StandardWorldTime)
		err := c.levyActionService.ExecuteCronLevyActions(ctx, worldTime)
		if err != nil {
			arg := db.CreateWorldTimeRecordParams{
				StopReason:     err.Error(),
				WorldStoppedAt: worldTime,
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
