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
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	levyRouterInit     sync.Once
	levyRouterInstance *levyRouter
)

type levyRouter struct {
	levyService  *service.LevyService
	realmService *service.RealmService
}

func NewLevyRouter(router *Api) {
	levyRouterInit.Do(func() {
		levyRouterInstance = &levyRouter{
			levyService:  router.service.LevyService,
			realmService: router.service.RealmService,
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

	arg1 := db.GetRealmIdWithSectorParams{
		OwnerID:    authPayload.UserId,
		CellNumber: req.Encampment,
	}

	foundRealmSector, err := r.realmService.GetMyRealmIdFromSectorNumber(ctx, &arg1)
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

	if !foundRealmSector.CellNumber.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, fmt.Sprint(foundRealmSector.Name, "에 속하지 않은 영토입니다."), nil),
		})
		return
	}

	combatUnitCount := req.Swordmen + req.Archers + req.ShieldBearers + req.Lancers

	movementSpeed := float64(req.Swordmen*util.SwordmanSpeed+
		req.Archers*util.ArcherSpeed+
		req.ShieldBearers*util.ShieldBearerSpeed+
		req.Lancers*util.LancerSpeed+
		req.SupplyTroop*util.SupplyTroopSpeed) / float64(combatUnitCount+req.SupplyTroop)

	offensiveStrength := (req.Swordmen*util.SwordmanOffensive +
		req.Archers*util.ArcherOffensive +
		req.ShieldBearers*util.ShieldBearerOffensive +
		req.Lancers*util.LancerOffensive) / (combatUnitCount)

	defensiveStrength := (req.Swordmen*util.SwordmanDefensive +
		req.Archers*util.ArcherDefensive +
		req.ShieldBearers*util.ShieldBearerDefensive +
		req.Lancers*util.LancerDefensive) / (combatUnitCount)

	arg2 := db.CreateLevyParams{
		RealmMemberID:     authPayload.UserId,
		MovementSpeed:     movementSpeed,
		OffensiveStrength: offensiveStrength,
		DefensiveStrength: defensiveStrength,
		Name:              req.Name,
		Morale:            util.DefaultMorale,
		Encampment:        req.Encampment,
		Swordmen:          req.Swordmen,
		Archers:           req.Archers,
		ShieldBearers:     req.ShieldBearers,
		Lancers:           req.Lancers,
		SupplyTroop:       req.SupplyTroop,
	}

	levy, stateCoffers, err := r.levyService.FormAUnit(ctx, foundRealmSector.RealmID, &arg2)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnprocessableEntity, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "국고가 부족합니다.", nil),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.CreateLevyResponse{
		APIResponse:  types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		StateCoffers: stateCoffers,
		Levy:         types.ToLevyResponse(levy),
		RealmMemberIDs: &types.RealmMemberIDs{
			RealmID: foundRealmSector.RealmID,
			UserID:  authPayload.UserId,
		},
	})
}
