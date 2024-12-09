package cmd

import (
	"database/sql"
	"log"

	"github.com/byeoru/kania/api"
	"github.com/byeoru/kania/config"
	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/service"
)

type Cmd struct {
	config  *config.Config
	api     *api.Api
	service *service.Service
}

func NewCmd(filePath string) *Cmd {
	c := &Cmd{
		config: config.NewConfig(filePath),
	}

	conn, err := sql.Open(c.config.Database.DatabaseDriver, c.config.Database.DatabaseSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	c.service = service.NewService(store)
	c.api = api.NewApi(c.service)
	c.api.ServerStart(c.config.Server.Port)
	return c
}
