package api

import (
	"github.com/gin-gonic/gin"
	"simple-bank/internal/db"
)

type Server struct {
	store  *db.Store
	engine *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	server.engine = gin.Default()

	server.engine.POST("/accounts", server.createAccount)
	server.engine.GET("/accounts/:id", server.getAccount)
	server.engine.GET("/accounts", server.listAccounts)

	return server
}

func (s *Server) Start(address string) error {
	return s.engine.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
