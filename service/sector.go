package service

import (
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/gin-gonic/gin"
)

var (
	sectorServiceInit     sync.Once
	sectorServiceInstance *SectorService
)

type SectorService struct {
	store db.Store
}

func newSectorService(store db.Store) *SectorService {
	sectorServiceInit.Do(func() {
		sectorServiceInstance = &SectorService{
			store,
		}
	})
	return sectorServiceInstance
}

func (s *SectorService) IncreasePopulation(ctx *gin.Context, arg *db.UpdateCensusPopulationParams) error {
	return s.store.UpdateCensusPopulation(ctx, arg)
}

func (s *SectorService) ApplyCensus(ctx *gin.Context, realmArg *db.UpdateCensusAtParams, sectorArg *db.UpdateCensusPopulationParams) error {
	return s.store.ExecTx(ctx, func(q *db.Queries) error {
		err := q.UpdateCensusPopulation(ctx, sectorArg)
		if err != nil {
			return err
		}

		err = q.UpdateCensusAt(ctx, realmArg)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *SectorService) GetPopulationAndCheck(ctx *gin.Context, cellNumber int32, userId int64) (int32, bool, error) {
	var population int32
	var isOwner bool
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		result, err := q.GetPopulation(ctx, cellNumber)
		if err != nil {
			return err
		}
		arg := db.CheckCellOwnerParams{
			RealmID: result.RealmID,
			OwnerID: userId,
		}
		ok, err := q.CheckCellOwner(ctx, &arg)
		if err != nil {
			return err
		}
		population = result.Population
		isOwner = ok
		return nil
	})
	return population, isOwner, err
}
