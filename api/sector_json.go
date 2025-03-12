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
	sectorJsonRouterInit     sync.Once
	sectorJsonRouterInstance *sectorJsonRouter
)

type sectorJsonRouter struct {
	realmMemberService *service.RealmMemberService
	levyActionService  *service.LevyActionService
	sectorJsonService  *service.SectorJsonService
}

func NewSectorJsonRouter(router *API) {
	sectorJsonRouterInit.Do(func() {
		sectorJsonRouterInstance = &sectorJsonRouter{
			realmMemberService: router.service.RealmMemberService,
			levyActionService:  router.service.LevyActionService,
			sectorJsonService:  router.service.SectorJsonService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/sector_jsons", sectorJsonRouterInstance.getBothJsons)
}

func (r *sectorJsonRouter) getBothJsons(ctx *gin.Context) {
	var reqQuery types.GetBothJsonsQueryRequest
	if err := ctx.ShouldBindUri(&reqQuery); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.GetBothJsonsResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindFullRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetBothJsonsResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetBothJsonsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmMember.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.FindMyLevyResponse{
			APIResponse: types.NewAPIResponse(false, "유저가 소속된 국가가 존재하지 않습니다.", nil),
		})
		return
	}

	action, err := r.levyActionService.FindOne(ctx, reqQuery.ActionID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetBothJsonsResponse{
				APIResponse: types.NewAPIResponse(false, "명령 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetBothJsonsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !(action.RealmID == me.RealmMember.RealmID.Int64 ||
		(action.TargetRealmID.Valid && (action.TargetRealmID.Int64 == me.RealmMember.RealmID.Int64))) {
		ctx.JSON(http.StatusForbidden, &types.GetBothJsonsResponse{
			APIResponse: types.NewAPIResponse(false, "해당 명령 정보에 접근 권한이 없습니다.", nil),
		})
	}

	arg := db.FindBothJsonbParams{
		RealmID1: action.TargetRealmID.Int64,
		RealmID2: action.RealmID,
	}
	jsons, err := r.sectorJsonService.FindBoth(ctx, &arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetBothJsonsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetBothJsonsResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Attacker: &types.JsonbResponse{
			RealmID: jsons[0].RealmSectorsJsonbID,
			Jsonb:   jsons[0].CellsJsonb,
		},
		Defender: &types.JsonbResponse{
			RealmID: jsons[1].RealmSectorsJsonbID,
			Jsonb:   jsons[1].CellsJsonb,
		},
	})
}
