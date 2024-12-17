package service

import (
	"encoding/json"
	"sync"

	db "github.com/byeoru/kania/db/repository"
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

func (s *RealmService) FindMyRealms(ctx *gin.Context, userId int64) (db.FindRealmWithJsonRow, error) {
	return s.store.FindRealmWithJson(ctx, userId)
}

func (s *RealmService) RegisterRealm(
	ctx *gin.Context,
	realm *db.CreateRealmParams,
	sector *db.CreateSectorParams,
) error {
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		realmId, err := q.CreateRealm(ctx, *realm)
		if err != nil {
			return err
		}
		sector.RealmID = realmId
		err = q.CreateSector(ctx, *sector)
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
		err = q.CreateRealmSectorsJsonb(ctx, db.CreateRealmSectorsJsonbParams{
			RealmID:    realmId,
			CellsJsonb: json,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
