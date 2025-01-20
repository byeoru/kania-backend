package cron

import (
	"context"
	"sync"
	"time"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
)

var (
	LevyActionCronInit     sync.Once
	LevyActionCronInstance *LevyActionCron
)

type LevyActionCron struct {
	levyActionService      *service.LevyActionService
	worldTimeRecordService *service.WorldTimeRecordService
}

func NewLevyActionCron(cron *Cron) *LevyActionCron {
	LevyActionCronInit.Do(func() {
		LevyActionCronInstance = &LevyActionCron{
			levyActionService: cron.service.LevyActionService,
		}
	})
	return LevyActionCronInstance
}

func (c *LevyActionCron) ExecuteLevyAction(ctx *context.Context, worldTime time.Time) {
	err := c.levyActionService.ExecuteCronLevyActions(ctx, worldTime)
	if err != nil {
		arg := db.CreateWorldTimeRecordParams{
			StopReason:     err.Error(),
			WorldStoppedAt: worldTime,
		}
		c.worldTimeRecordService.CreateWorldTimeRecord(ctx, &arg)
	}
}
