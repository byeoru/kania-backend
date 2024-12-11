package api

import (
	"net/http"
	"sync"

	"github.com/byeoru/kania/service"
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
	router.engine.GET("/realms", realmRouterInstance.getMyRealms)
}

func (r *realmRouter) getMyRealms(ctx *gin.Context) {
	var req types.GetMyRealmsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.SignupUserResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	realms, err := r.realmService.FindAllMyRealms(ctx, req.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetMyRealmsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetMyRealmsResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Realms:      types.ToRealmsResponse(realms),
	})
}
