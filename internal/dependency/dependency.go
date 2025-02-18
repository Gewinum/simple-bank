package dependency

import (
	"database/sql"
	_ "github.com/lib/pq"
	"go.uber.org/dig"
	"simple-bank/internal/api"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
	"simple-bank/internal/tokens"
	"simple-bank/internal/utils"
)

func NewDependency() *dig.Container {
	container := dig.New()

	utils.NoError(container.Provide(newConfig))
	utils.NoError(container.Provide(newPasetoManager))
	utils.NoError(container.Provide(newSqlConnection))
	utils.NoError(container.Provide(db.NewStore))
	utils.NoError(container.Provide(api.NewServer))

	return container
}

func newConfig() (*config.Config, error) {
	return config.Load("./")
}

func newPasetoManager(cfg *config.Config) (tokens.Manager, error) {
	return tokens.NewPasetoManager(cfg.TokenPrivateKey)
}

func newSqlConnection(cfg *config.Config) (*sql.DB, error) {
	return sql.Open(cfg.DBDriver, cfg.DBSource)
}
