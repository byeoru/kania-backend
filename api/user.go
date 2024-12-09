package api

import (
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
		router.engine.POST("/signup", userRouterInstance.create)
	})
}

func (u *userRouter) create(ctx *gin.Context) {
	var req types.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
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

	if err := u.userService.CreateUser(ctx, &arg); err != nil {
		pqErr, ok := err.(*pq.Error)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if pqErr.Code.Name() == "unique_violation" {
			switch pqErr.Constraint {
			case "users_email_key":
				ctx.JSON(http.StatusForbidden, pqErr.Message)
			case "users_nickname_key":
				ctx.JSON(http.StatusForbidden, pqErr.Message)
			}
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
