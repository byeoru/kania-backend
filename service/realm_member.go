package service

import (
	"database/sql"
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
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.GetRealmId(ctx, id)
}
