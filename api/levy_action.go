package api

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	errors "github.com/byeoru/kania/types/error"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	levyActionRouterInit     sync.Once
	levyActionRouterInstance *levyActionRouter
)

type levyActionRouter struct {
	levyActionService *service.LevyActionService
	levyService       *service.LevyService
	sectorService     *service.SectorService
}

func NewLevyActionRouter(router *API) {
	levyActionRouterInit.Do(func() {
		levyActionRouterInstance = &levyActionRouter{
			sectorService:     router.service.SectorService,
			levyService:       router.service.LevyService,
			levyActionService: router.service.LevyActionService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.POST("/api/levy_action/advance", levyActionRouterInstance.advance)
}

func (r *levyActionRouter) advance(ctx *gin.Context) {
	var reqJson types.AttackJsonRequest
	if err := ctx.ShouldBindJSON(&reqJson); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}
	var reqQuery types.AttackQueryRequest
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg1 := db.FindLevyActionCountByLevyIdParams{
		LevyID:        reqQuery.LevyID,
		ReferenceDate: reqJson.CurrentWorldTime,
	}

	levyActionCount, err := r.levyActionService.FindLevyActionByLevyId(ctx, &arg1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if levyActionCount != 0 {
		ctx.JSON(http.StatusUnprocessableEntity, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "이미 이동중인 부대입니다.", nil),
		})
		return
	}

	err = r.sectorService.CheckOriginTargetSectorValid(ctx, authPayload.UserId, reqJson.OriginSector, reqJson.TargetSector)
	if err != nil {
		if txError, ok := err.(*errors.TxError); ok {
			ctx.JSON(txError.Code, &types.AttackResponse{
				APIResponse: types.NewAPIResponse(false, txError.Message, txError.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	bMyLevy, err := r.levyService.IsMyLevy(ctx, authPayload.UserId, reqQuery.LevyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.AttackResponse{
				APIResponse: types.NewAPIResponse(false, "존재하지 않는 부대입니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !bMyLevy {
		ctx.JSON(http.StatusForbidden, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "해당 부대에 대한 권한이 없습니다.", nil),
		})
		return
	}

	arg2 := db.CreateLevyActionParams{
		LevyID:       reqQuery.LevyID,
		OriginSector: reqJson.OriginSector,
		TargetSector: reqJson.TargetSector,
		ActionType:   util.Attack,
		Completed:    false,
		StartedAt:    reqJson.StartedAt,
		// NOTE: 현재 날짜로 test중
		ExpectedCompletionAt: reqJson.CurrentWorldTime,
	}

	err = r.levyActionService.ExecuteLevyAction(ctx, &arg2)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.AttackResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
	})
}
