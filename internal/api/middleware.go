package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	tokens2 "simple-bank/internal/tokens"
	"strings"
)

const (
	authorizationHeader     = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "payload"
)

func getPayloadFromGinCtx(c *gin.Context) *tokens2.Payload {
	return c.MustGet(authorizationPayloadKey).(*tokens2.Payload)
}

func authMiddleware(tokensManager tokens2.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationToken := c.Request.Header.Get(authorizationHeader)
		if authorizationToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("authorization header is missing")))
			return
		}

		fields := strings.Fields(authorizationToken)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(errors.New("authorization header is invalid")))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("unsupported authorization type: %s", authorizationType)))
			return
		}

		accessToken := fields[1]
		payload, err := tokensManager.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("invalid token: %s", err.Error())))
			return
		}

		c.Set(authorizationPayloadKey, payload)
		c.Next()
	}
}
