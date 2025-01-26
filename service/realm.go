package service

import (
	"database/sql"
	"encoding/json"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	realmServiceInit     sync.Once
	realmServiceInstance *RealmService
)

type RealmService struct {
	store db.Store
}

func newRealmService(store db.Store) *RealmService {
	realmServiceInit.Do(func() {
		realmServiceInstance = &RealmService{
			store,
		}
	})
	return realmServiceInstance
}

func (s *RealmService) RegisterRealm(
	ctx *gin.Context,
	realmArg *db.CreateRealmParams,
	sectorArg *db.CreateSectorParams,
	userId int64,
) (*db.Realm, error) {
	var result *db.Realm
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		realmArg.Color = realmArg.Color[1:]
		realm, err := q.CreateRealm(ctx, realmArg)
		if err != nil {
			return err
		}

		err = q.CreateRealmMember(ctx, &db.CreateRealmMemberParams{
			RmID:         sql.NullInt64{Int64: userId, Valid: true},
			RealmID:      sql.NullInt64{Int64: realm.RealmID, Valid: true},
			Status:       util.Chief,
			PrivateMoney: util.DefaultPrivateMoney,
		})
		if err != nil {
			return err
		}

		arg1 := db.CreateMemberAuthorityParams{
			RmID:          userId,
			CreateUnit:    true,
			ReinforceUnit: true,
			MoveUnit:      true,
			AttackUnit:    true,
			PrivateTroops: true,
			Census:        true,
		}
		err = q.CreateMemberAuthority(ctx, &arg1)
		if err != nil {
			return err
		}

		sectorArg.RealmID = realm.RealmID
		err = q.CreateSector(ctx, sectorArg)
		if err != nil {
			return err
		}

		arg2 := db.AddCapitalParams{
			RealmID: realm.RealmID,
			Capital: sectorArg.CellNumber,
		}
		err = q.AddCapital(ctx, &arg2)
		if err != nil {
			return err
		}
		realm.Capitals = append(realm.Capitals, sectorArg.CellNumber)

		json, err := json.Marshal(map[int32]int32{
			sectorArg.CellNumber: sectorArg.CellNumber,
		})
		if err != nil {
			return err
		}
		err = q.CreateRealmSectorsJsonb(ctx, &db.CreateRealmSectorsJsonbParams{
			RealmSectorsJsonbID: realm.RealmID,
			CellsJsonb:          json,
		})
		if err != nil {
			return err
		}
		result = realm
		return nil
	})
	return result, err
}

func (s *RealmService) GetDataForCensus(ctx *gin.Context, realmId int64) (*db.GetCensusAndPopulationGrowthRateRow, error) {
	return s.store.GetCensusAndPopulationGrowthRate(ctx, realmId)
}

func (s *RealmService) AddCapital(ctx *gin.Context, arg *db.AddCapitalParams) error {
	return s.store.AddCapital(ctx, arg)
}

func (s *RealmService) GetRealmOwnerRmId(ctx *gin.Context, realmId int64) (int64, error) {
	return s.store.GetRealmOwnerRmId(ctx, realmId)
}
