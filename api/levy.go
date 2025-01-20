package api

import (
	"database/sql"
	"fmt"
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
	levyRouterInit     sync.Once
	levyRouterInstance *levyRouter
)

type levyRouter struct {
	levyService        *service.LevyService
	realmService       *service.RealmService
	realmMemberService *service.RealmMemberService
}

func NewLevyRouter(router *API) {
	levyRouterInit.Do(func() {
		levyRouterInstance = &levyRouter{
			levyService:        router.service.LevyService,
			realmService:       router.service.RealmService,
			realmMemberService: router.service.RealmMemberService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.POST("/api/levies", levyRouterInstance.createLevy)
}

func (r *levyRouter) createLevy(ctx *gin.Context) {
	var req types.CreateLevyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg1 := db.GetMyRmIdOfSectorParams{
		UserID:     sql.NullInt64{Int64: authPayload.UserId, Valid: true},
		CellNumber: req.Encampment,
	}
	rmId, err := r.realmMemberService.GetMyRmIdOfSector(ctx, &arg1)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "해당 군에 대한 권한이 없습니다.", nil),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "서버 오류입니다. CODE: 1", err.Error()),
		})
		return
	}

	arg2 := db.GetRealmIdWithSectorParams{
		RmID:       rmId,
		CellNumber: req.Encampment,
	}

	foundRealmSector, err := r.realmService.GetMyRealmIdFromSectorNumber(ctx, &arg2)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "소속된 국가가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "서버 오류입니다. CODE: 2", err.Error()),
		})
		return
	}

	if !foundRealmSector.CellNumber.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, fmt.Sprint(foundRealmSector.Name, "에 속하지 않은 영토입니다."), nil),
		})
		return
	}

	movementSpeed := util.CalculateLevyAdvanceSpeed(&db.Levy{
		Swordmen:      req.Swordmen,
		Archers:       req.Archers,
		ShieldBearers: req.ShieldBearers,
		Lancers:       req.Lancers,
		SupplyTroop:   req.SupplyTroop,
	})

	arg3 := db.CreateLevyParams{
		RmID:          rmId,
		MovementSpeed: movementSpeed,
		Name:          req.Name,
		Morale:        util.DefaultMorale,
		Encampment:    req.Encampment,
		Swordmen:      req.Swordmen,
		Archers:       req.Archers,
		ShieldBearers: req.ShieldBearers,
		Lancers:       req.Lancers,
		SupplyTroop:   req.SupplyTroop,
		Stationed:     true,
	}

	levy, resultInfo, err := r.levyService.FormAUnit(ctx, foundRealmSector.RealmID, &arg3)
	if err != nil {
		if txError, ok := err.(*errors.TxError); ok {
			ctx.JSON(txError.Code, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, txError.Message, txError.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "서버 오류입니다. CODE: 3", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.CreateLevyResponse{
		APIResponse:  types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		StateCoffers: resultInfo.StateCoffers,
		Population:   resultInfo.Population,
		Levy:         types.ToLevyResponse(levy),
		LevyAffiliation: &types.LevyAffiliation{
			RealmID: foundRealmSector.RealmID,
			RmID:    rmId,
		},
	})
}
