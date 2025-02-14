package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
	"simple-bank/internal/config"
	"testing"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	cfg, err := config.Load("../../")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("can't connect to db", err)
	}

	testDB = conn
	testQueries = New(conn)

	os.Exit(m.Run())
}
