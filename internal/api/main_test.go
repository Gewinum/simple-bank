package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
	"simple-bank/internal/random"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	cfg := config.Config{
		TokenPrivateKey:     random.String(32),
		AccessTokenDuration: time.Minute * 15,
	}

	server, err := NewServer(&cfg, store)
	require.NoError(t, err)
	
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
