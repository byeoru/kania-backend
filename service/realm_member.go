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

func (s *RealmMemberService) FindRealmMember(ctx *gin.Context, userId int64) (*db.RealmMember, error) {
	return s.store.FindRealmMember(ctx, userId)
}

func (s *RealmMemberService) FindFullRealmMember(ctx *gin.Context, userId int64) (*db.FindFullRealmMemberRow, error) {
	return s.store.FindFullRealmMember(ctx, userId)
}
