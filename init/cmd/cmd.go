package cmd

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/byeoru/kania/api"
	"github.com/byeoru/kania/config"
	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
	"github.com/byeoru/kania/token"
)

type Cmd struct {
	api     *api.Api
	service *service.Service
}

func NewCmd(filePath string) *Cmd {
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

	c.service = service.NewService(store)
	c.api = api.NewApi(c.service)
	c.api.ServerStart(config.Server.Port)
	return c
}
