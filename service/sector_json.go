package service

import (
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/gin-gonic/gin"
)

var (
	sectorJsonServiceInit     sync.Once
	sectorJsonServiceInstance *SectorJsonService
)

type SectorJsonService struct {
	store db.Store
}

func newSectorJsonService(store db.Store) *SectorJsonService {
	sectorJsonServiceInit.Do(func() {
		sectorJsonServiceInstance = &SectorJsonService{
			store,
		}
	})
	return sectorJsonServiceInstance
}

func (s *SectorJsonService) FindBoth(ctx *gin.Context, arg *db.FindBothJsonbParams) ([]*db.RealmSectorsJsonb, error) {
	return s.store.FindBothJsonb(ctx, arg)
}
