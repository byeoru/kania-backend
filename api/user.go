package api

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/byeoru/kania/config"
	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/types"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

var (
	userRouterInit     sync.Once
	userRouterInstance *userRouter
)

type userRouter struct {
	userService *service.UserService
}

func newUserRouter(router *API) {
	userRouterInit.Do(func() {
		userRouterInstance = &userRouter{
			userService: router.service.UserService,
		}
		router.engine.POST("/api/signup", userRouterInstance.signup)
		router.engine.POST("/api/login", userRouterInstance.login)
	})
}

func (r *userRouter) signup(ctx *gin.Context) {
	var req types.SignupUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.SignupUserResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	hashedPassword, err := r.userService.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	arg := db.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		Nickname:       req.Nickname,
	}

	if err := r.userService.Signup(ctx, &arg); err != nil {
		pqErr, ok := err.(*pq.Error)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, &types.SignupUserResponse{
				APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
			})
			return
		}

		if pqErr.Code.Name() == "unique_violation" {
			switch pqErr.Constraint {
			case "users_email_key":
				ctx.JSON(http.StatusConflict, &types.SignupUserResponse{
					APIResponse: types.NewAPIResponse(false, "이미 사용 중인 이메일입니다.", pqErr.Detail),
				})
			case "users_nickname_key":
				ctx.JSON(http.StatusConflict, &types.SignupUserResponse{
					APIResponse: types.NewAPIResponse(false, "이미 사용 중인 닉네임입니다.", pqErr.Detail),
				})
			}
			return
		}
	}

	ctx.JSON(http.StatusOK, &types.SignupUserResponse{
		APIResponse: types.NewAPIResponse(true, "회원가입이 완료되었습니다.", nil),
	})
}

func (r *userRouter) login(ctx *gin.Context) {
	var req types.LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	user, err := r.userService.Login(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, &types.LoginUserResponse{
				APIResponse: types.NewAPIResponse(false, "존재하지 않는 계정입니다.", sql.ErrNoRows.Error()),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	err = r.userService.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "존재하지 않는 계정입니다.", sql.ErrNoRows.Error()),
		})
		return
	}

	token, err := r.userService.CreateToken(user.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	// Set the SameSite mode to Strict.
	ctx.SetSameSite(http.SameSiteStrictMode)

	// 쿠키 설정
	ctx.SetCookie(
		config.GetInstance().Cookie.AccessCookieName, // 쿠키 이름
		token, // 쿠키 값
		config.GetInstance().Cookie.CookieDuration, // 쿠키 만료 시간 (초 단위) - 1시간
		"/",         // 쿠키 유효 경로
		"localhost", // 쿠키 도메인 (테스트 시 localhost)
		false,       // Secure: HTTPS에서만 전송 (테스트 시 false)
		true,        // HttpOnly: JavaScript 접근 차단
	)

	ctx.JSON(http.StatusOK, &types.LoginUserResponse{
		APIResponse: types.NewAPIResponse(true, "로그인이 완료되었습니다.", nil),
	})
}
