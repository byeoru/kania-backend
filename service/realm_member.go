package service

import (
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/gin-gonic/gin"
)

var (
	realmMemberServiceInit     sync.Once
	realmMemberServiceInstance *RealmMemberService
)

type RealmMemberService struct {
	store db.Store
}

func newRealmMemberService(store db.Store) *RealmMemberService {
	realmMemberServiceInit.Do(func() {
		realmMemberServiceInstance = &RealmMemberService{
			store,
		}
	})
	return realmMemberServiceInstance
}

func (s *RealmMemberService) GetMyRealmId(ctx *gin.Context, userId int64) (int64, error) {
	return s.store.GetRealmId(ctx, userId)
}

func (s *RealmMemberService) GetMyRmIdOfSector(ctx *gin.Context, arg *db.GetMyRmIdOfSectorParams) (int64, error) {
	return s.store.GetMyRmIdOfSector(ctx, arg)
}
