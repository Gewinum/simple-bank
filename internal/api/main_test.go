package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"os"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
	"simple-bank/internal/random"
	"simple-bank/internal/tokens"
	"testing"
	"time"
)

func newTestContainer(t *testing.T, store db.Store) *dig.Container {
	container := dig.New()

	require.NoError(t, container.Provide(getFakeConfig))
	require.NoError(t, container.Provide(wrapDatabase(store)))
	require.NoError(t, container.Provide(getPasetoManager))
	require.NoError(t, container.Provide(NewServer))

	return container
}

func getFakeConfig() *config.Config {
	return &config.Config{
		TokenPrivateKey:     random.String(32),
		AccessTokenDuration: time.Minute * 15,
	}
}

// wrapDatabase is done like this to allow mock db values
func wrapDatabase(store db.Store) func() db.Store {
	return func() db.Store {
		return store
	}
}

func getPasetoManager(cfg *config.Config) (tokens.Manager, error) {
	return tokens.NewPasetoManager(cfg.TokenPrivateKey)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
