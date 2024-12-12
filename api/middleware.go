package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "access token이 존재하지 않습니다.", err.Error()))
			return
		}

		fields := strings.Fields(authorizationHeader)
		// 접두사 (ex Bearer) 가 없는 경우
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "유효하지 않은 authorization format입니다.", err.Error()))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "지원하지 않은 authorization type입니다.", err.Error()))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, types.NewAPIResponse(false, "사용할 수 없는 access token입니다.", err.Error()))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
