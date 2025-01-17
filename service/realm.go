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

func (s *RealmService) FindMyRealm(ctx *gin.Context, userId int64) (*db.FindRealmWithJsonRow, error) {
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.FindRealmWithJson(ctx, id)
}

func (s *RealmService) FindAllRealmExcludeMe(ctx *gin.Context, userId int64) ([]*db.FindAllRealmsWithJsonExcludeMeRow, error) {
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.FindAllRealmsWithJsonExcludeMe(ctx, id)
}

func (s *RealmService) RegisterRealm(
	ctx *gin.Context,
	realmArg *db.CreateRealmParams,
	sectorArg *db.CreateSectorParams,
) (*db.Realm, error) {
	var result *db.Realm
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		realmArg.Color = realmArg.Color[1:]
		realm, err := q.CreateRealm(ctx, realmArg)
		if err != nil {
			return err
		}

		sectorArg.RealmID = realm.RealmID
		err = q.CreateSector(ctx, sectorArg)
		if err != nil {
			return err
		}

		arg := db.AddCapitalParams{
			RealmID: realm.RealmID,
			Capital: sectorArg.CellNumber,
		}
		err = q.AddCapital(ctx, &arg)
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
		err = q.CreateRealmMember(ctx, &db.CreateRealmMemberParams{
			UserID:       realm.OwnerID.Int64,
			Status:       util.Chief,
			PrivateMoney: util.DefaultPrivateMoney,
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

func (s *RealmService) GetMyRealmId(ctx *gin.Context, userId int64) (int64, error) {
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.GetRealmId(ctx, id)
}

func (s *RealmService) GetMyRealmIdFromSectorNumber(ctx *gin.Context, arg *db.GetRealmIdWithSectorParams) (*db.GetRealmIdWithSectorRow, error) {
	return s.store.GetRealmIdWithSector(ctx, arg)
}

func (s *RealmService) GetOurRealmLevies(ctx *gin.Context, realmId int64) ([]*db.GetOurRealmLeviesRow, error) {
	return s.store.GetOurRealmLevies(ctx, realmId)
}

func (s *RealmService) AddCapital(ctx *gin.Context, arg *db.AddCapitalParams) error {
	return s.store.AddCapital(ctx, arg)
}
