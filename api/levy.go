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
	levyRouterInit     sync.Once
	levyRouterInstance *levyRouter
)

type levyRouter struct {
	levyService        *service.LevyService
	realmService       *service.RealmService
	realmMemberService *service.RealmMemberService
	sectorService      *service.SectorService
}

func NewLevyRouter(router *API) {
	levyRouterInit.Do(func() {
		levyRouterInstance = &levyRouter{
			levyService:        router.service.LevyService,
			realmService:       router.service.RealmService,
			realmMemberService: router.service.RealmMemberService,
			sectorService:      router.service.SectorService,
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

	me, err := r.realmMemberService.FindFullRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmMember.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.GetIndigenousUnitResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	bOurSector, err := r.sectorService.IsOurSector(ctx, req.Encampment, me.RealmMember.RealmID.Int64)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "권한이 없는 영토입니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !bOurSector {
		ctx.JSON(http.StatusForbidden, &types.CreateLevyResponse{
			APIResponse: types.NewAPIResponse(false, "권한이 없는 영토입니다.", nil),
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

	levyRmId := me.MemberAuthority.RmID
	// 부대 소유권한이 없을 경우
	if !me.MemberAuthority.PrivateTroops {
		ownerRmId, err := r.realmService.GetRealmOwnerRmId(ctx, me.RealmMember.RealmID.Int64)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, &types.CreateLevyResponse{
					APIResponse: types.NewAPIResponse(false, "국가 정보가 존재하지 않습니다.", err.Error()),
				})
				return
			}
			ctx.JSON(http.StatusInternalServerError, &types.CreateLevyResponse{
				APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
			})
			return
		}
		levyRmId = ownerRmId
	}

	arg3 := db.CreateLevyParams{
		RmID:          levyRmId,
		RealmID:       sql.NullInt64{Int64: me.RealmMember.RealmID.Int64, Valid: true},
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

	levy, resultInfo, err := r.levyService.FormAUnit(ctx, me.RealmMember.RealmID.Int64, &arg3)
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
			RealmID: me.RealmMember.RealmID.Int64,
			RmID:    levyRmId,
		},
	})
}
