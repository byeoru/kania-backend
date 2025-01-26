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

func (s *SectorService) GetPopulationAndCheck(ctx *gin.Context, sector int32, realmId int64) (int32, error) {
	r, err := s.store.GetPopulation(ctx, sector)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.NewTxError(http.StatusUnprocessableEntity, "소유 중인 영토가 아닙니다.")
		}
		return 0, err
	}

	if r.RealmID != realmId {
		return 0, errors.NewTxError(http.StatusForbidden, "다른 국가의 영토입니다.")
	}

	return r.Population, nil
}

func (s *SectorService) CheckOriginTargetSectorValidForAttack(ctx *gin.Context, realmId int64, targetSector int32) error {
	targetSectorRealmId, err := s.store.GetSectorRealmId(ctx, targetSector)
	if err != nil {
		// 누구의 지배도 받지 않는 땅인 경우 Pass
		if err != sql.ErrNoRows {
			return err
		}
	}
	// 공격할 섹터가 나의 영토인지 확인
	if targetSectorRealmId == realmId {
		return errors.NewTxError(http.StatusUnprocessableEntity, "자신의 영토는 공격할 수 없습니다.")
	}
	return nil
}

func (s *SectorService) CheckOriginTargetSectorValidForMove(ctx *gin.Context, realmId int64, targetSector int32) error {
	targetSectorRealmId, err := s.store.GetSectorRealmId(ctx, targetSector)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewTxError(http.StatusUnprocessableEntity, "자신의 영토 내에서만 이동할 수 있습니다.")
		}
		return err
	}
	// 목적지가 섹터가 나의 영토인지 확인
	if targetSectorRealmId != realmId {
		return errors.NewTxError(http.StatusUnprocessableEntity, "자신의 영토 내에서만 이동할 수 있습니다.")
	}
	return nil
}

func (s *SectorService) IsOurSector(ctx *gin.Context, sector int32, realmId int64) (bool, error) {
	sectorRealmId, err := s.store.GetSectorRealmId(ctx, sector)
	return realmId == sectorRealmId, err
}
