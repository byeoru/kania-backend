package api

import (
	"github.com/byeoru/kania/service"
	"github.com/gin-gonic/gin"
)

type Api struct {
	engine  *gin.Engine
	service *service.Service
}

func NewApi(service *service.Service) *Api {
	r := &Api{
		engine:  gin.New(),
		service: service,
	}

	// router
	newUserRouter(r)
	NewRealmRouter(r)
	return r
}

func (a *Api) ServerStart(port string) error {
	return a.engine.Run(port)
}
