package service

import (
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

func (s *RealmService) FindAllMyRealms(ctx *gin.Context, userId int64) ([]db.Realm, error) {
	return s.store.FindAllRealms(ctx, userId)
}

func (s *RealmService) RegisterRealmWithSector(ctx *gin.Context, realm *db.CreateRealmParams, sector *db.CreateSectorParams) error {
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
		return nil
	})
	return err
}
