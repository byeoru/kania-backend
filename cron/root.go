package cron

import (
	"sync"

	"github.com/byeoru/kania/service"
)

var (
	cronInit     sync.Once
	cronInstance *Cron
)

type Cron struct {
	LevyActionCron *LevyActionCron
}

func NewCron(s *service.Service) *Cron {
	cronInit.Do(func() {
		cronInstance = &Cron{
			LevyActionCron: NewLevyActionCron(s),
		}
	})
	return cronInstance
}
