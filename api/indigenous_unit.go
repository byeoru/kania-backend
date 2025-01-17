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
	indigenousUnitRouterInit     sync.Once
	indigenousUnitRouterInstance *indigenousUnitRouter
)

type indigenousUnitRouter struct {
	indigenousUnitService *service.IndigenousUnitService
}

func NewIndigenousUnitRouter(router *Api) {
	indigenousUnitRouterInit.Do(func() {
		indigenousUnitRouterInstance = &indigenousUnitRouter{
			indigenousUnitService: router.service.IndigenousUnitService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/indigenous_units/:sector_number", indigenousUnitRouterInstance.GetIndigenousUnit)
}

func (r *indigenousUnitRouter) GetIndigenousUnit(ctx *gin.Context) {
	var reqPath types.GetIndigenousUnitPathRequest
	if err := ctx.ShouldBindUri(&reqPath); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	unit, err := r.indigenousUnitService.FindIndigenousUnit(ctx, reqPath.SectorNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.GetIndigenousUnitResponse{
				APIResponse: types.NewAPIResponse(false, "토착 세력이 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetIndigenousUnitResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetIndigenousUnitResponse{
		APIResponse:    types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		IndigenousUnit: types.NewIndigenousUnitResponse(unit),
	})
}
