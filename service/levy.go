package service

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/types"
	errors "github.com/byeoru/kania/types/error"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	levyServiceInit     sync.Once
	levyServiceInstance *LevyService
)

type LevyService struct {
	store db.Store
}

func newLevyService(store db.Store) *LevyService {
	levyServiceInit.Do(func() {
		levyServiceInstance = &LevyService{
			store,
		}
	})
	return levyServiceInstance
}

func (s *LevyService) FormAUnit(ctx *gin.Context, myRealmId int64, levyArg *db.CreateLevyParams) (*db.Levy, *types.CreateLevyResultInfo, error) {
	var levy *db.Levy
	var resultInfo *types.CreateLevyResultInfo
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		unitStat := util.GetUnitStat()
		wholeProductionCost := levyArg.Swordmen*unitStat.Swordman.ProductionCost +
			levyArg.Archers*unitStat.Archer.ProductionCost +
			levyArg.ShieldBearers*unitStat.ShieldBearer.ProductionCost +
			levyArg.Lancers*unitStat.Lancer.ProductionCost +
			levyArg.SupplyTroop*unitStat.SupplyTroop.ProductionCost

		updatedCoffers, err := s.store.UpdateStateCoffers(ctx, &db.UpdateStateCoffersParams{
			RealmID:   myRealmId,
			Deduction: wholeProductionCost,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.NewTxError(http.StatusUnprocessableEntity, "국고가 부족합니다.")
			}
			return err
		}

		newLevy, err := q.CreateLevy(ctx, levyArg)
		if err != nil {
			return err
		}

		population, err := q.UpdatePopulation(ctx, &db.UpdatePopulationParams{
			Cellnumber: newLevy.Encampment,
			Deduction:  levyArg.Swordmen + levyArg.Archers + levyArg.ShieldBearers + levyArg.Lancers + levyArg.SupplyTroop,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.NewTxError(http.StatusUnprocessableEntity, "인구가 부족합니다.")
			}
			return err
		}

		levy = newLevy
		resultInfo = &types.CreateLevyResultInfo{
			StateCoffers: updatedCoffers,
			Population:   population,
		}
		return nil
	})
	return levy, resultInfo, err
}

func (s *LevyService) FindLevyInfoWithAuthority(ctx *gin.Context, levyId int64) (*db.FindLevyInfoWithAuthorityRow, error) {
	return s.store.FindLevyInfoWithAuthority(ctx, levyId)
}

func (s *LevyService) FindOurRealmLevies(ctx *gin.Context, realmId int64) ([]*db.Levy, error) {
	id := sql.NullInt64{Int64: realmId, Valid: true}
	return s.store.FindOurRealmLevies(ctx, id)
}

func (s *LevyService) FindLevy(ctx *gin.Context, levyId int64) (*db.Levy, error) {
	return s.store.FindLevy(ctx, levyId)
}

func (s *LevyService) FindEncampmentLevies(ctx *gin.Context, arg *db.FindEncampmentLeviesParams) ([]*db.Levy, error) {
	return s.store.FindEncampmentLevies(ctx, arg)
}
