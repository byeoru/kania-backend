package api

import (
	"database/sql"
	"net/http"
	"sync"

	db "github.com/byeoru/kania/db/repository"
	grpcclient "github.com/byeoru/kania/grpc_client"
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
	levyActionService  *service.LevyActionService
	levyService        *service.LevyService
	sectorService      *service.SectorService
	realmMemberService *service.RealmMemberService
	rpcClient          *grpcclient.Client
}

func NewLevyActionRouter(router *API) {
	levyActionRouterInit.Do(func() {
		levyActionRouterInstance = &levyActionRouter{
			sectorService:      router.service.SectorService,
			levyService:        router.service.LevyService,
			levyActionService:  router.service.LevyActionService,
			realmMemberService: router.service.RealmMemberService,
			rpcClient:          router.grpcClient,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.POST("/api/levy_actions/advance", levyActionRouterInstance.advance)
	authRoutes.POST("/api/levy_actions/move", levyActionRouterInstance.move)
	authRoutes.GET("/api/levy_actions", levyActionRouterInstance.findLevyAction)
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

	me, err := r.realmMemberService.FindRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.AttackResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	levyOwnership, err := r.levyService.FindLevyInfoWithAuthority(ctx, reqQuery.LevyID)
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

	if levyOwnership.RealmID != me.RealmID {
		ctx.JSON(http.StatusForbidden, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "다른 국가의 군부대에 명령을 내릴 수 없습니다.", nil),
		})
		return
	}

	if !levyOwnership.AttackUnit {
		ctx.JSON(http.StatusForbidden, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "공격 명령을 내릴 수 있는 권한이 없습니다.", nil),
		})
		return
	}

	currentWorldTime := util.CalculateCurrentWorldTime(util.StandardRealTime, util.StandardWorldTime)
	arg1 := db.FindLevyActionCountByLevyIdParams{
		LevyID:        reqQuery.LevyID,
		ReferenceDate: currentWorldTime,
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

	err = r.sectorService.CheckOriginTargetSectorValidForAttack(ctx, me.RealmID.Int64, reqJson.TargetSector)
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

	distance, err := r.rpcClient.GetDistance(levyOwnership.Encampment, reqJson.TargetSector)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	totalDurationHours := distance / levyOwnership.MovementSpeed
	cd := util.CalculateCompletionDate(currentWorldTime, totalDurationHours)

	arg3 := db.CreateLevyActionParams{
		LevyID:               reqQuery.LevyID,
		RealmID:              me.RealmID.Int64,
		RmID:                 me.RmID,
		OriginSector:         levyOwnership.Encampment,
		TargetSector:         reqJson.TargetSector,
		Distance:             distance,
		ActionType:           util.Attack,
		Completed:            false,
		StartedAt:            currentWorldTime,
		ExpectedCompletionAt: cd,
	}

	levyAction, err := r.levyActionService.ExecuteLevyAction(ctx, &arg3)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.AttackResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.AttackResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		LevyAction:  types.ToLevyActionResponse(levyAction),
	})
}

func (r *levyActionRouter) move(ctx *gin.Context) {
	var reqJson types.MoveJsonRequest
	if err := ctx.ShouldBindJSON(&reqJson); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}
	var reqQuery types.MoveQueryRequest
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.MoveResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "소속된 국가가 없습니다.", nil),
		})
		return
	}

	levyOwnership, err := r.levyService.FindLevyInfoWithAuthority(ctx, reqQuery.LevyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.MoveResponse{
				APIResponse: types.NewAPIResponse(false, "존재하지 않는 부대입니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if levyOwnership.RealmID != me.RealmID {
		ctx.JSON(http.StatusForbidden, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "다른 국가의 군부대에 명령을 내릴 수 없습니다.", nil),
		})
		return
	}

	if !levyOwnership.MoveUnit {
		ctx.JSON(http.StatusForbidden, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "주둔지 이동 명령을 내릴 수 있는 권한이 없습니다.", nil),
		})
		return
	}

	currentWorldTime := util.CalculateCurrentWorldTime(util.StandardRealTime, util.StandardWorldTime)
	arg1 := db.FindLevyActionCountByLevyIdParams{
		LevyID:        reqQuery.LevyID,
		ReferenceDate: currentWorldTime,
	}

	levyActionCount, err := r.levyActionService.FindLevyActionByLevyId(ctx, &arg1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if levyActionCount != 0 {
		ctx.JSON(http.StatusUnprocessableEntity, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "이미 이동중인 부대입니다.", nil),
		})
		return
	}

	err = r.sectorService.CheckOriginTargetSectorValidForMove(ctx, me.RealmID.Int64, reqJson.TargetSector)
	if err != nil {
		if txError, ok := err.(*errors.TxError); ok {
			ctx.JSON(txError.Code, &types.MoveResponse{
				APIResponse: types.NewAPIResponse(false, txError.Message, txError.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	distance, err := r.rpcClient.GetDistance(levyOwnership.Encampment, reqJson.TargetSector)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	totalDurationHours := distance / levyOwnership.MovementSpeed
	cd := util.CalculateCompletionDate(currentWorldTime, totalDurationHours)

	arg3 := db.CreateLevyActionParams{
		LevyID:               reqQuery.LevyID,
		RealmID:              me.RealmID.Int64,
		RmID:                 me.RmID,
		OriginSector:         levyOwnership.Encampment,
		TargetSector:         reqJson.TargetSector,
		ActionType:           util.Move,
		Completed:            false,
		StartedAt:            currentWorldTime,
		ExpectedCompletionAt: cd,
	}

	levyAction, err := r.levyActionService.ExecuteLevyAction(ctx, &arg3)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.MoveResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.MoveResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		LevyAction:  types.ToLevyActionResponse(levyAction),
	})
}

func (r *levyActionRouter) findLevyAction(ctx *gin.Context) {
	var reqQuery types.FindLevyActionQueryRequest
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.FindLevyActionResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	me, err := r.realmMemberService.FindFullRealmMember(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.FindLevyActionResponse{
				APIResponse: types.NewAPIResponse(false, "유저 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.FindLevyActionResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	action, err := r.levyActionService.FindLevyAction(ctx, reqQuery.LevyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.FindLevyActionResponse{
				APIResponse: types.NewAPIResponse(false, "부대 명령 정보가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.FindLevyActionResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	if !me.RealmMember.RealmID.Valid {
		ctx.JSON(http.StatusUnprocessableEntity, &types.FindLevyActionResponse{
			APIResponse: types.NewAPIResponse(false, "유저가 소속된 국가가 존재하지 않습니다.", nil),
		})
		return
	}

	if me.RealmMember.RealmID.Int64 != action.RealmID {
		ctx.JSON(http.StatusForbidden, &types.FindLevyActionResponse{
			APIResponse: types.NewAPIResponse(false, "다른 국가의 부대 명령 정보는 확인할 수 없습니다.", nil),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.FindLevyActionResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		LevyAction:  types.ToLevyActionResponse(action),
	})
}
