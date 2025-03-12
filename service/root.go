package service

import (
	"sync"

	db "github.com/byeoru/kania/db/repository"
	grpcclient "github.com/byeoru/kania/grpc_client"
)

var (
	serviceInit     sync.Once
	serviceInstance *Service
)

type Service struct {
	UserService            *UserService
	RealmService           *RealmService
	SectorService          *SectorService
	RealmMemberService     *RealmMemberService
	LevyService            *LevyService
	LevyActionService      *LevyActionService
	IndigenousUnitService  *IndigenousUnitService
	WorldTimeRecordService *WorldTimeRecordService
	SectorJsonService      *SectorJsonService
}

func NewService(store db.Store, grpcClient *grpcclient.Client) *Service {
	serviceInit.Do(func() {
		serviceInstance = &Service{
			UserService:            newUserService(store),
			RealmService:           newRealmService(store),
			SectorService:          newSectorService(store),
			RealmMemberService:     newRealmMemberService(store),
			LevyService:            newLevyService(store),
			LevyActionService:      newLevyActionService(store, grpcClient),
			IndigenousUnitService:  newIndigenousUnitService(store),
			WorldTimeRecordService: newWorldTimeRecordService(store),
			SectorJsonService:      newSectorJsonService(store),
		}
	})
	return serviceInstance
}
