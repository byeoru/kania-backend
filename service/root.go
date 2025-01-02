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
	UserService        *UserService
	RealmService       *RealmService
	SectorService      *SectorService
	RealmMemberService *RealmMemberService
	LevyService        *LevyService
}

func NewService(store db.Store) *Service {
	serviceInit.Do(func() {
		serviceInstance = &Service{
			UserService:        newUserService(store),
			RealmService:       newRealmService(store),
			SectorService:      newSectorService(store),
			RealmMemberService: newRealmMemberService(store),
			LevyService:        newLevyService(store),
		}
	})
	return serviceInstance
}
