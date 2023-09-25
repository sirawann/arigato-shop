package myTests

import (
	"encoding/json"

	"github.com/sirawann/arigato-shop/config"
	"github.com/sirawann/arigato-shop/modules/servers"
	databases "github.com/sirawann/arigato-shop/pkg/databases"
)

func SetupTest() servers.IModuleFactory {
	cfg := config.LoadConfig("../.env.test")

	db := databases.DbConnect(cfg.Db())

	s := servers.NewSever(cfg, db)
	return servers.InitModule(nil, s.GetServer(), nil)
}

func CompressToJSON(obj any) string {
	result, _ := json.Marshal(&obj)
	return string(result)
}