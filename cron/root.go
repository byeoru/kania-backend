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
	service *service.Service
}

func NewCron(service *service.Service) *Cron {
	cronInit.Do(func() {
		cronInstance = &Cron{
			service: service,
		}
		NewLevyActionCron(cronInstance)
	})
	return cronInstance
}
