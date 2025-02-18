package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"simple-bank/internal/config"
	"simple-bank/internal/db"
	"simple-bank/tokens"
)

type Server struct {
	config        *config.Config
	store         db.Store
	engine        *gin.Engine
	tokensManager tokens.Manager
}

func NewServer(config *config.Config, store db.Store) (*Server, error) {
	tokensManager, err := tokens.NewPasetoManager(config.TokenPrivateKey)
	if err != nil {
		return nil, err
	}
	server := &Server{
		config:        config,
		store:         store,
		engine:        gin.Default(),
		tokensManager: tokensManager,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("currency", validateCurrency)
		if err != nil {
			return nil, err
		}
	}

	server.engine.POST("/users", server.createUser)
	server.engine.POST("/users/login", server.loginUser)

	authRoutes := server.engine.Group("/").Use(authMiddleware(server.tokensManager))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccounts)

	authRoutes.POST("/transfers", server.createTransfer)

	return server, nil
}

func (s *Server) Start(address string) error {
	return s.engine.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
