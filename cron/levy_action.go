package cron

import (
	"context"
	"sync"
	"time"

	"github.com/byeoru/kania/service"
)

var (
	LevyActionCronInit     sync.Once
	LevyActionCronInstance *LevyActionCron
)

type LevyActionCron struct {
	levyActionService *service.LevyActionService
}

func NewLevyActionCron(cron *Cron) *LevyActionCron {
	LevyActionCronInit.Do(func() {
		LevyActionCronInstance = &LevyActionCron{
			levyActionService: cron.service.LevyActionService,
		}
	})
	return LevyActionCronInstance
}

func (c *LevyActionCron) ExecuteLevyAction(ctx context.Context, worldTime time.Time) {

}
