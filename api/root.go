package api

import (
	"context"
	"log"
	"net/http"

	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/types"
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

	// politicalEntity validator
	// if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	// 	v.RegisterValidation("politicalEntity", types.ValidPoliticalEntity)
	// } else {
	// 	log.Fatal("politicalEntity validator setting error")
	// }

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("hexColor", types.ValidColor)
	} else {
		log.Fatal("hexColor validator setting error")
	}

	// router
	newUserRouter(r)
	NewRealmRouter(r)
	NewSectorRouter(r)
	NewLevyRouter(r)
	NewRealmMemberRouter(r)
	NewLevyActionRouter(r)
	NewIndigenousUnitRouter(r)

	// 토착 세력 병력 초기화
	ctx := context.Background()
	err := r.service.IndigenousUnitService.InitIndigenousUnit(&ctx)
	if err != nil {
		log.Fatalf("Failed to initialize indigenous unit: %v", err)
	}

	r.engine.LoadHTMLFiles("static/index.html")
	r.engine.Static("/assets", "static/assets")
	r.engine.Static("/public/assets", "static/assets")

	r.engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.engine.GET("/world", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	return r
}

func (a *Api) ServerStart(port string) error {
	return a.engine.Run(port)
}
