package api

import (
	"log"

	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

	// cors
	r.engine.Use(cors.New(
		cors.Config{
			AllowOrigins: []string{"http://localhost:5173"},
		}))

	// politicalEntity validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("politicalEntity", types.ValidPoliticalEntity)
	} else {
		log.Fatal("validator setting error")
	}

	// router
	newUserRouter(r)
	NewRealmRouter(r)
	return r
}

func (a *Api) ServerStart(port string) error {
	return a.engine.Run(port)
}
