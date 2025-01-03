package api

import (
	"database/sql"
	"net/http"
	"sync"
	"time"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	realmRouterInit     sync.Once
	realmRouterInstance *realmRouter
)

type realmRouter struct {
	realmService       *service.RealmService
	userService        *service.UserService
	sectorService      *service.SectorService
	realmMemberService *service.RealmMemberService
}

func NewRealmRouter(router *Api) {
	realmRouterInit.Do(func() {
		realmRouterInstance = &realmRouter{
			realmService:       router.service.RealmService,
			userService:        router.service.UserService,
			sectorService:      router.service.SectorService,
			realmMemberService: router.service.RealmMemberService,
		}
	})
	authRoutes := router.engine.Group("/").Use(authMiddleware(token.GetTokenMakerInstance()))
	authRoutes.GET("/api/realms/me", realmRouterInstance.getMyRealm)
	authRoutes.GET("/api/realms", realmRouterInstance.getMeAndOthersReams)
	authRoutes.POST("/api/realms", realmRouterInstance.establishARealm)
	authRoutes.POST("/api/realms/census", realmRouterInstance.executeCensus)
}

func (r *realmRouter) getMyRealm(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	realm, err := r.realmService.FindMyRealm(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetMyRealmsResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetMyRealmsResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		Realm:       types.ToMyRealmResponse(realm),
	})
}

func (r *realmRouter) getMeAndOthersReams(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	myRealm, err := r.realmService.FindMyRealm(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNoContent, gin.H{})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.GetMeAndOthersReams{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	theOthersRealms, err := r.realmService.FindAllRealmExcludeMe(ctx, authPayload.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.GetMeAndOthersReams{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.GetMeAndOthersReams{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		MyRealm:     types.ToMyRealmResponse(myRealm),
		TheOthersRealms: types.Map(theOthersRealms, func(realm *db.FindAllRealmsWithJsonExcludeMeRow) *types.RealmResponse {
			return types.ToTheOthersRealmsResponse(realm)
		}),
	})
}

func (r *realmRouter) establishARealm(ctx *gin.Context) {
	var req types.EstablishARealmRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	owner, err := r.userService.FindUser(ctx, authPayload.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	realmArg := db.CreateRealmParams{
		Name:                 req.Name,
		OwnerID:              authPayload.UserId,
		OwnerNickname:        owner.Nickname,
		CapitalNumber:        req.CellNumber,
		PoliticalEntity:      "Tribe",
		Color:                req.RealmColor,
		PopulationGrowthRate: util.TribePopulationGrowthRate,
		StateCoffers:         util.DefaultStateCoffers,
		CensusAt:             req.InitDate,
		TaxCollectionAt:      req.InitDate,
	}
	sectorArg := db.CreateSectorParams{
		CellNumber:     req.CellNumber,
		ProvinceNumber: req.ProvinceNumber,
		Population:     req.Population,
	}

	result, err := r.realmService.RegisterRealm(ctx, &realmArg, &sectorArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.EstablishARealmResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.EstablishARealmResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
		MyRealm:     types.ToMyRealmFromEntityResponse(result),
	})
}

func (r *realmRouter) executeCensus(ctx *gin.Context) {
	var req types.ExecuteCensusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.ExecuteCensusResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// 속한 국가의 realm id를 가져옴
	myRealmId, err := r.realmMemberService.GetMyRealmId(ctx, authPayload.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, &types.ExecuteCensusResponse{
				APIResponse: types.NewAPIResponse(false, "소속된 국가가 존재하지 않습니다.", err.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.ExecuteCensusResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	censusMetadata, err := r.realmService.GetDataForCensus(ctx, myRealmId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.ExecuteCensusResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	// 두 날짜의 차이 계산
	duration := req.CurrentDate.Sub(censusMetadata.CensusAt)
	days := duration.Hours() / 24
	// 이전 인구조사 실시 후 1년이 지났는지 확인 (1년 = 365.25일, 평균적으로)
	oneYear := 31557600
	if duration <= time.Duration(oneYear)*time.Second {
		ctx.JSON(http.StatusTooManyRequests, &types.ExecuteCensusResponse{
			APIResponse: types.NewAPIResponse(false, "1년 이상의 기간이 지나지 않았습니다.", nil),
		})
		return
	}

	realmArg := db.UpdateCensusAtParams{
		RealmID:  myRealmId,
		CensusAt: req.CurrentDate,
	}
	sectorArg := db.UpdateCensusPopulationParams{
		RealmID:        myRealmId,
		DurationDay:    days,
		RateOfIncrease: censusMetadata.PopulationGrowthRate,
	}

	err = r.sectorService.ApplyCensus(ctx, &realmArg, &sectorArg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.ExecuteCensusResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.ExecuteCensusResponse{
		APIResponse: types.NewAPIResponse(true, "요청이 성공적으로 완료되었습니다.", nil),
	})
}
