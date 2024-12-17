package api

import (
	"net/http"

	"github.com/byeoru/kania/config"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/types"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authorization
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationCookie, err := ctx.Cookie(config.GetInstance().Cookie.AccessCookieName)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "Cookie가 존재하지 않습니다.", err.Error()))
			return
		}

		payload, err := tokenMaker.VerifyToken(authorizationCookie)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "사용할 수 없는 access token입니다.", err.Error()))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
