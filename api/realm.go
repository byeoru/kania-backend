package api

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/gin-gonic/gin"
)

var (
	realmRouterInit     sync.Once
	realmRouterInstance *realmRouter
)

type realmRouter struct {
	realmService *service.RealmService
}

func NewRealmRouter(router *Api) {
	realmRouterInit.Do(func() {
		realmRouterInstance = &realmRouter{
			realmService: router.service.RealmService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/realms", realmRouterInstance.getMyRealm)
	authRoutes.POST("/realms", realmRouterInstance.establishARealm)
}

func (r *realmRouter) getMyRealm(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	realm, err := r.realmService.FindMyRealms(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetMyRealmsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetMyRealmsResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Realms:      types.ToRealmsResponse(realm),
	})
}

func (r *realmRouter) establishARealm(ctx *gin.Context) {
	var req types.EstablishARealmRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	realmArg := db.CreateRealmParams{
		Name:            req.Name,
		OwnerID:         authPayload.UserId,
		CapitalNumber:   req.CellNumber,
		PoliticalEntity: req.PoliticalEntity,
	}
	sectorArg := db.CreateSectorParams{
		CellNumber:     req.CellNumber,
		ProvinceNumber: req.ProvinceNumber,
	}

	err := r.realmService.RegisterRealm(ctx, &realmArg, &sectorArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.EstablishARealmResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
	})
}
