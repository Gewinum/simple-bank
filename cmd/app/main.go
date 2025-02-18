package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"simple-bank/api"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
)

func main() {
	cfg, err := config.Load("./")
	if err != nil {
		panic(err)
	}
	sqlConn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		panic(err)
	}

	store := db.NewStore(sqlConn)
	server, err := api.NewServer(cfg, store)
	if err != nil {
		panic(err)
	}
	err = server.Start(cfg.ServerAddress)
	if err != nil {
		panic(err)
	}
}
