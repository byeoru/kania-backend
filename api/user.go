package api

import (
	"database/sql"
	"net/http"
	"sync"

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

func newUserRouter(router *Api) {
	userRouterInit.Do(func() {
		userRouterInstance = &userRouter{
			userService: router.service.UserService,
		}
		router.engine.POST("/signup", userRouterInstance.signup)
		router.engine.POST("/login", userRouterInstance.login)
	})
}

func (u *userRouter) signup(ctx *gin.Context) {
	var req types.SignupUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.SignupUserResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	hashedPassword, err := u.userService.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	arg := db.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		Nickname:       req.Nickname,
	}

	if err := u.userService.Signup(ctx, &arg); err != nil {
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
				ctx.JSON(http.StatusForbidden, &types.SignupUserResponse{
					APIResponse: types.NewAPIResponse(false, "이미 사용 중인 이메일입니다.", pqErr.Detail),
				})
			case "users_nickname_key":
				ctx.JSON(http.StatusForbidden, &types.SignupUserResponse{
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

func (u *userRouter) login(ctx *gin.Context) {
	var req types.LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "올바르지 않은 요청 데이터입니다.", err.Error()),
		})
		return
	}

	user, err := u.userService.Login(ctx, req.Email)
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

	err = u.userService.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "존재하지 않는 계정입니다.", sql.ErrNoRows.Error()),
		})
		return
	}

	token, err := u.userService.CreateToken(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &types.LoginUserResponse{
			APIResponse: types.NewAPIResponse(false, "알 수 없는 오류입니다.", err.Error()),
		})
		return
	}

	ctx.JSON(http.StatusOK, &types.LoginUserResponse{
		APIResponse: types.NewAPIResponse(true, "로그인이 완료되었습니다.", nil),
		AccessToken: token,
	})
}
