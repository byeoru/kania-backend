package api

import (
	"sync"

	"github.com/byeoru/kania/service"
)

var (
	realmMemberRouterInit     sync.Once
	realmMemberRouterInstance *realmMemberRouter
)

type realmMemberRouter struct {
	realmMemberService *service.RealmMemberService
	realmService       *service.RealmService
}

func NewRealmMemberRouter(router *API) {
	realmMemberRouterInit.Do(func() {
		realmMemberRouterInstance = &realmMemberRouter{
			realmMemberService: router.service.RealmMemberService,
			realmService:       router.service.RealmService,
		}
	})
	// authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
}
