package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"simple-bank/api"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:root@localhost:5432/simplebank?sslmode=disable"
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
	server := api.NewServer(store)
	err = server.Start(cfg.ServerAddress)
	if err != nil {
		panic(err)
	}
}
