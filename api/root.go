package api

import (
	"log"
	"net/http"

	grpcclient "github.com/byeoru/kania/grpc_client"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type API struct {
	engine     *gin.Engine
	service    *service.Service
	grpcClient *grpcclient.Client
}

func NewApi(service *service.Service, grpcClient *grpcclient.Client) *API {
	r := &API{
		engine:     gin.New(),
		service:    service,
		grpcClient: grpcClient,
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

	// CORS 미들웨어 설정
	r.engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		}, // 허용할 오리진 (프론트엔드 도메인)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},            // 허용할 HTTP 메서드
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // 허용할 헤더
		AllowCredentials: true,                                                // 쿠키를 사용할 수 있도록 허용
		ExposeHeaders:    []string{"Content-Length", "Authorization"},         // 클라이언트에서 접근할 수 있는 헤더
		AllowWildcard:    false,                                               // 모든 도메인 허용 여부 (기본값 false)
	}))

	// router
	newUserRouter(r)
	NewRealmRouter(r)
	NewSectorRouter(r)
	NewLevyRouter(r)
	NewRealmMemberRouter(r)
	NewLevyActionRouter(r)
	NewIndigenousUnitRouter(r)
	NewSectorJsonRouter(r)

	// 토착 세력 병력 초기화
	// ctx := context.Background()
	// err := r.service.IndigenousUnitService.InitIndigenousUnit(&ctx)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize indigenous unit: %v", err)
	// }

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

func (a *API) ServerStart(port string) error {
	return a.engine.Run(port)
}
