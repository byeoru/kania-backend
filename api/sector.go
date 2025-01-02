package api

import (
	"net/http"
	"sync"

	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/gin-gonic/gin"
)

var (
	sectorRouterInit     sync.Once
	sectorRouterInstance *SectorRouter
)

type SectorRouter struct {
	sectorService *service.SectorService
}

func NewSectorRouter(router *Api) {
	sectorRouterInit.Do(func() {
		sectorRouterInstance = &SectorRouter{
			sectorService: router.service.SectorService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/sectors/:cell_number/population", sectorRouterInstance.getPopulation)
}

func (r *SectorRouter) getPopulation(ctx *gin.Context) {
	var req types.GetPopulationRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	pop, ok, err := r.sectorService.GetPopulationAndCheck(ctx, req.CellNumber, authPayload.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusForbidden, &types.GetPopulationResponse{
			APIResponse: types.NewAPIResponse(false, "해당 섹터에 대한 권한이 없습니다.", nil),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetPopulationResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Population:  pop,
	})
}
