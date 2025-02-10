package api

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	errors "github.com/byeoru/kania/types/error"
	"github.com/gin-gonic/gin"
)

var (
	sectorRouterInit     sync.Once
	sectorRouterInstance *SectorRouter
)

type SectorRouter struct {
	sectorService      *service.SectorService
	realmMemberService *service.RealmMemberService
}

func NewSectorRouter(router *API) {
	sectorRouterInit.Do(func() {
		sectorRouterInstance = &SectorRouter{
			sectorService:      router.service.SectorService,
			realmMemberService: router.service.RealmMemberService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/sectors/:cell_number/population", sectorRouterInstance.getPopulation)
}

func (r *SectorRouter) getPopulation(ctx *gin.Context) {
	var req types.GetPopulationRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindFullRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetPopulationResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmMember.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	pop, err := r.sectorService.GetPopulationAndCheck(ctx, req.CellNumber, me.RealmMember.RealmID.Int64)
	if err != nil {
		if txErr, ok := err.(*errors.TxError); ok {
			ctx.JSON(txErr.Code, &types.GetPopulationResponse{
				APIResponse: types.NewAPIResponse(false, txErr.Message, txErr.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetPopulationResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Population:  pop,
	})
}
