package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/byeoru/kania/api"
	"github.com/byeoru/kania/config"
	"github.com/byeoru/kania/cron"
	db "github.com/byeoru/kania/db/repository"
	grpcclient "github.com/byeoru/kania/grpc_client"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
	"github.com/byeoru/kania/util"
)

type Cmd struct {
	api     *api.API
	service *service.Service
	cron    *cron.Cron
}

func NewCmd(filePath string) *Cmd {
	// grpc client 생성
	grpcClient := grpcclient.NewClient()

	c := new(Cmd)
	config.LoadConfig(filePath)
	config := config.GetInstance()

	conn, err := sql.Open(config.Database.DatabaseDriver, config.Database.DatabaseSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	err = token.InitPasetoMaker(config.Token.TokenSymmertricKey)
	if err != nil {
		err = fmt.Errorf("cannot create token maker: %w", err)
		log.Fatal(err)
	}

	c.service = service.NewService(store, grpcClient)
	c.api = api.NewApi(c.service, grpcClient)
	c.cron = cron.NewCron(c.service)

	/**
	cron 실행
	*/
	ticker := time.NewTicker(time.Second)
	ctx := context.Background()
	go c.cron.LevyActionCron.ExecuteCronLevyActions(&ctx, ticker)

	go c.api.ServerStart(config.Server.Port)

	// 종료 신호 채널 생성
	quit := make(chan os.Signal, 1)
	// SIGINT (Ctrl+C), SIGTERM 신호를 채널에 전달
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 종료 신호 대기
	<-quit
	fmt.Println("\nShutting down server...")

	// 종료 전에 실행할 함수 호출
	ticker.Stop()
	currentWorldTime := util.CalculateCurrentWorldTime(util.StandardRealTime, util.StandardWorldTime)
	c.cron.LevyActionCron.RecordWorldTime(&ctx, currentWorldTime)
	conn.Close()
	grpcClient.Conn.Close()

	fmt.Println("Server exiting")
	return c
}
