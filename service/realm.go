package service

import (
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
	return s.store.FindRealmWithJson(ctx, userId)
}

func (s *RealmService) FindAllRealmExcludeMe(ctx *gin.Context, userId int64) ([]*db.FindAllRealmsWithJsonExcludeMeRow, error) {
	return s.store.FindAllRealmsWithJsonExcludeMe(ctx, userId)
}

func (s *RealmService) RegisterRealm(
	ctx *gin.Context,
	realm *db.CreateRealmParams,
	sector *db.CreateSectorParams,
) (*db.Realm, error) {
	var result *db.Realm
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		realm.Color = realm.Color[1:]
		realm, err := q.CreateRealm(ctx, realm)
		if err != nil {
			return err
		}

		sector.RealmID = realm.RealmID
		err = q.CreateSector(ctx, sector)
		if err != nil {
			return err
		}
		// TODO: 나중에 struct 만들기
		json, err := json.Marshal(gin.H{
			"cells": []int32{sector.CellNumber},
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
			RealmID:      realm.RealmID,
			UserID:       realm.OwnerID,
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
	return s.store.GetRealmId(ctx, userId)
}

func (s *RealmService) GetMyRealmIdFromSectorNumber(ctx *gin.Context, arg *db.GetRealmIdWithSectorParams) (*db.GetRealmIdWithSectorRow, error) {
	return s.store.GetRealmIdWithSector(ctx, arg)
}
