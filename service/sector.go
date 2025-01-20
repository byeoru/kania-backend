package service

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	errors "github.com/byeoru/kania/types/error"
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

func (s *SectorService) GetPopulationAndCheck(ctx *gin.Context, cellNumber int32, rmId int64) (int32, bool, error) {
	var population int32
	var isOwner bool
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		result, err := q.GetPopulation(ctx, cellNumber)
		if err != nil {
			return err
		}
		arg := db.CheckCellOwnerParams{
			RealmID: result.RealmID,
			RmID:    rmId,
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

func (s *SectorService) CheckOriginTargetSectorValid(ctx *gin.Context, rmId int64, originSector int32, targetSector int32) error {
	myRealmId, err := s.store.GetRealmIdByRmId(ctx, rmId)
	if err != nil {
		return err
	}
	originSectorRealmId, err := s.store.GetSectorRealmId(ctx, originSector)
	if err != nil {
		return err
	}
	if !myRealmId.Valid {
		return errors.NewTxError(http.StatusUnprocessableEntity, "부대가 소속된 국가 정보가 없습니다.")
	}
	// 출발지가 나의 영토인지 확인
	if originSectorRealmId != myRealmId.Int64 {
		return errors.NewTxError(http.StatusUnprocessableEntity, "출발지가 나의 영토가 아닙니다.")
	}
	targetSectorRealmId, err := s.store.GetSectorRealmId(ctx, targetSector)
	if err != nil {
		// 누구의 지배도 받지 않는 땅인 경우 Pass
		if err != sql.ErrNoRows {
			return err
		}
	}
	// 공격할 섹터가 나의 영토인지 확인
	if targetSectorRealmId == myRealmId.Int64 {
		return errors.NewTxError(http.StatusUnprocessableEntity, "자신의 영토는 공격할 수 없습니다.")
	}
	return nil
}
