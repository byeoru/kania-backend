package api

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/gin-gonic/gin"
)

var (
	realmMemberRouterInit     sync.Once
	realmMemberRouterInstance *realmMemberRouter
)

type realmMemberRouter struct {
	realmMemberService *service.RealmMemberService
	realmService       *service.RealmService
}

func NewRealmMemberRouter(router *Api) {
	realmMemberRouterInit.Do(func() {
		realmMemberRouterInstance = &realmMemberRouter{
			realmMemberService: router.service.RealmMemberService,
			realmService:       router.service.RealmService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/realm_members/levies", realmMemberRouterInstance.getRealmMembersLevies)
}

func (r *realmMemberRouter) getRealmMembersLevies(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	myRealmId, err := r.realmService.GetMyRealmId(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "소속된 국가가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	membersAndLevies, err := r.realmMemberService.GetRealmMembersLevies(ctx, myRealmId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetRealmMembersLeviesResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetRealmMembersLeviesResponse{
		APIResponse:  types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		RealmMembers: types.ToRealmMembers(membersAndLevies),
	})
}
