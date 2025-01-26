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

func (s *RealmMemberService) FindRealmMember(ctx *gin.Context, userId int64) (*db.RealmMember, error) {
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.FindRealmMember(ctx, id)
}

func (s *RealmMemberService) FindFullRealmMember(ctx *gin.Context, userId int64) (*db.FindFullRealmMemberRow, error) {
	id := sql.NullInt64{Int64: userId, Valid: true}
	return s.store.FindFullRealmMember(ctx, id)
}

func (s *RealmMemberService) FindMyRealm(ctx *gin.Context, realmId int64) (*db.FindRealmWithJsonRow, error) {
	return s.store.FindRealmWithJson(ctx, realmId)
}

func (s *RealmMemberService) FindAllRealmExcludeMe(ctx *gin.Context, realmId int64) ([]*db.FindAllRealmsWithJsonExcludeMeRow, error) {
	return s.store.FindAllRealmsWithJsonExcludeMe(ctx, realmId)
}
