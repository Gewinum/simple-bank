package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	tokens2 "simple-bank/internal/tokens"
	"testing"
	"time"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokensManager tokens2.Manager,
	authorizationType string,
	payloadCreationParams tokens2.PayloadCreationParams,
) {
	token, err := tokensManager.CreateToken(payloadCreationParams)
	require.NoError(t, err)

	authorizationToken := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Add(authorizationHeader, authorizationToken)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenManager tokens2.Manager)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, tokens2.PayloadCreationParams{
					Subject:   "user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "no_authorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "unsupported_authorization_type",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {
				addAuthorization(t, request, tokenManager, "unsupported", tokens2.PayloadCreationParams{
					Subject:   "user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "invalid_authorization_type",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {
				addAuthorization(t, request, tokenManager, "", tokens2.PayloadCreationParams{
					Subject:   "user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  time.Minute,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "expired_token",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, tokens2.PayloadCreationParams{
					Subject:   "user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now(),
					Duration:  -time.Minute,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "using_token_early",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager tokens2.Manager) {
				addAuthorization(t, request, tokenManager, authorizationTypeBearer, tokens2.PayloadCreationParams{
					Subject:   "user",
					Audience:  "test",
					Issuer:    "test",
					NotBefore: time.Now().Add(time.Minute),
					Duration:  time.Minute,
				})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testContainer := newTestContainer(t, nil)

			require.NoError(t, testContainer.Invoke(func(server *Server) {
				authPath := "/auth"
				server.engine.GET(
					authPath,
					authMiddleware(server.tokensManager),
					func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{})
					},
				)

				recorder := httptest.NewRecorder()
				request, err := http.NewRequest(http.MethodGet, authPath, nil)
				require.NoError(t, err)

				tc.setupAuth(t, request, server.tokensManager)
				server.engine.ServeHTTP(recorder, request)
				tc.checkResponse(t, recorder)
			}))
		})
	}
}
