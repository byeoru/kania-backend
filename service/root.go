package service

import (
	"sync"

	db "github.com/byeoru/kania/db/repository"
)

var (
	serviceInit     sync.Once
	serviceInstance *Service
)

type Service struct {
	UserService *UserService
}

func NewService(store db.Store) *Service {
	serviceInit.Do(func() {
		serviceInstance = &Service{
			UserService: newUserService(store),
		}
	})
	return serviceInstance
}
