package service

import (
	"fmt"
	"sync"

	db "github.com/byeoru/kania/db/repository"
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

func (s *LevyService) FormAUnit(ctx *gin.Context, myRealmId int64, levyArg *db.CreateLevyParams) (*db.Levy, int32, error) {
	var levy *db.Levy
	var stateCoffers int32
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		wholeProductionCost := levyArg.Swordmen*util.SwordsmanProductionCost +
			levyArg.Archers*util.ArcherProductionCost +
			levyArg.ShieldBearers*util.ShieldBearerProductionCost +
			levyArg.Lancers*util.LancerProductionCost +
			levyArg.SupplyTroop*util.SupplyTroopProductionCost

		updatedCoffers, err := s.store.UpdateStateCoffers(ctx, &db.UpdateStateCoffersParams{
			RealmID:   myRealmId,
			Deduction: wholeProductionCost,
		})

		if err != nil {
			return err
		}

		newLevy, err := q.CreateLevy(ctx, levyArg)
		if err != nil {
			fmt.Println("2", err)
			return err
		}

		levy = newLevy
		stateCoffers = updatedCoffers
		return nil
	})
	return levy, stateCoffers, err
}
