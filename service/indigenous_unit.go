package service

import (
	"context"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/gin-gonic/gin"
)

var (
	indigenousUnitServiceInit     sync.Once
	indigenousUnitServiceInstance *IndigenousUnitService
)

type IndigenousUnitService struct {
	store db.Store
}

func newIndigenousUnitService(store db.Store) *IndigenousUnitService {
	indigenousUnitServiceInit.Do(func() {
		indigenousUnitServiceInstance = &IndigenousUnitService{
			store,
		}
	})
	return indigenousUnitServiceInstance
}

func (s *IndigenousUnitService) InitIndigenousUnit(ctx *context.Context) error {
	return s.store.InitIndigenousUnits(*ctx)
}

func (s *IndigenousUnitService) FindIndigenousUnit(ctx *gin.Context, sectorNumber int32) (*db.IndigenousUnit, error) {
	return s.store.FindIndigenousUnit(ctx, sectorNumber)
}
