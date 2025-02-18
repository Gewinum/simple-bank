package main

import (
	_ "github.com/lib/pq"
	"simple-bank/internal/api"
	"simple-bank/internal/config"
	"simple-bank/internal/dependency"
	"simple-bank/internal/utils"
)

func main() {
	dpd := dependency.NewDependency()

	utils.NoError(dpd.Invoke(func(server *api.Server, cfg *config.Config) {
		go func() {
			utils.NoError(server.Start(cfg.ServerAddress))
		}()
	}))

	select {}
}
