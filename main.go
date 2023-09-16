package main

import (
	_"fmt"
	"os"

	"github.com/sirawann/arigato-shop/config"
	"github.com/sirawann/arigato-shop/modules/servers"
	database "github.com/sirawann/arigato-shop/pkg/databases"
)

func envPath() string {
	if len(os.Args) == 1 {
		return ".env"
	} else {
		return os.Args[1]
	}
}

func main() {
	cfg := config.LoadConfig(envPath())

	db := database.DbConnect(cfg.Db())
	defer db.Close()

	servers.NewSever(cfg, db).Start()
}
