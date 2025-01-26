package api

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	realmMemberRouterInit     sync.Once
	realmMemberRouterInstance *realmMemberRouter
)

type realmMemberRouter struct {
	realmMemberService *service.RealmMemberService
	realmService       *service.RealmService
	levyService        *service.LevyService
}

func NewRealmMemberRouter(router *API) {
	realmMemberRouterInit.Do(func() {
		realmMemberRouterInstance = &realmMemberRouter{
			realmMemberService: router.service.RealmMemberService,
			realmService:       router.service.RealmService,
			levyService:        router.service.LevyService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/realm_members/realms", realmMemberRouterInstance.getMeAndOthersReams)
	authRoutes.GET("/api/realm_members/levies", realmMemberRouterInstance.getOurRealmLevies)
}

func (r *realmMemberRouter) getMeAndOthersReams(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetMeAndOthersReamsResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetMeAndOthersReamsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.GetMeAndOthersReamsResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	myRealms, err := r.realmMemberService.FindMyRealm(ctx, me.RealmID.Int64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetMeAndOthersReamsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	theOthersRealms, err := r.realmMemberService.FindAllRealmExcludeMe(ctx, me.RealmID.Int64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetMeAndOthersReamsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetMeAndOthersReamsResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		StandardTimes: &types.StandardTimes{
			StandardRealTime:  util.StandardRealTime,
			StandardWorldTime: util.StandardWorldTime,
		},
		MyRealm: types.ToMyRealmResponse(myRealms),
		TheOthersRealms: util.Map(theOthersRealms, func(realm *db.FindAllRealmsWithJsonExcludeMeRow) *types.RealmResponse {
			return types.ToTheOthersRealmsResponse(realm)
		}),
	})
}

func (r *realmMemberRouter) getOurRealmLevies(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindFullRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetRealmMembersLeviesResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetRealmMembersLeviesResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmMember.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.GetRealmMembersLeviesResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	levies, err := r.levyService.FindOurRealmLevies(ctx, me.RealmMember.RealmID.Int64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetRealmMembersLeviesResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetRealmMembersLeviesResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		RealmLevies: types.ToRealmLevies(levies),
	})
}
