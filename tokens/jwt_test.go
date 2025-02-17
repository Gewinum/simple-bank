package tokens

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"simple-bank/random"
	"testing"
	"time"
)

func TestJWTManager_Success(t *testing.T) {
	manager, err := NewJWTManager(random.String(32))
	require.NoError(t, err)

	username := random.Username()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := time.Now().Add(duration)

	token, err := manager.CreateToken(PayloadCreationParams{
		Subject:   username,
		Audience:  "bank-service",
		Issuer:    "test",
		NotBefore: time.Now(),
		Duration:  duration,
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := manager.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Subject)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager, err := NewJWTManager(random.String(32))
	require.NoError(t, err)
	username := random.Username()
	duration := -time.Minute
	token, err := manager.CreateToken(PayloadCreationParams{
		Subject:   username,
		Audience:  "bank-service",
		Issuer:    "test",
		NotBefore: time.Now(),
		Duration:  duration,
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)
	payload, err := manager.VerifyToken(token)
	require.ErrorIs(t, err, ErrTokenExpired)
	require.Empty(t, payload)
}

func TestJWTManager_SignatureInvalid(t *testing.T) {
	manager, err := NewJWTManager(random.String(32))
	require.NoError(t, err)
	username := random.Username()
	duration := time.Minute
	token, err := manager.CreateToken(PayloadCreationParams{
		Subject:   username,
		Audience:  "bank-service",
		Issuer:    "test",
		NotBefore: time.Now(),
		Duration:  duration,
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)
	manager2, err := NewJWTManager(random.String(32))
	require.NoError(t, err)
	payload, err := manager2.VerifyToken(token)
	require.ErrorIs(t, err, ErrTokenInvalid)
	require.Empty(t, payload)
}

func TestJWTManager_UsingTokenEarly(t *testing.T) {
	manager, err := NewJWTManager(random.String(32))
	require.NoError(t, err)
	username := random.Username()
	duration := time.Minute
	token, err := manager.CreateToken(PayloadCreationParams{
		Subject:   username,
		Audience:  "bank-service",
		Issuer:    "test",
		NotBefore: time.Now().Add(time.Second * 5),
		Duration:  duration,
	})
	require.NoError(t, err)
	require.NotEmpty(t, token)
	payload, err := manager.VerifyToken(token)
	require.ErrorIs(t, err, ErrTokenNotValidYet)
	require.Empty(t, payload)
}

func TestJWTManager_NoAlgorithm(t *testing.T) {
	manager, err := NewJWTManager(random.String(32))
	require.NoError(t, err)
	username := random.Username()
	duration := time.Minute
	payload, err := NewPayload(PayloadCreationParams{
		Subject:   username,
		Audience:  "bank-service",
		Issuer:    "test",
		NotBefore: time.Now(),
		Duration:  duration,
	})
	require.NoError(t, err)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	require.NotEmpty(t, jwtToken)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	payload2, err := manager.VerifyToken(token)
	require.ErrorIs(t, err, jwt.ErrTokenUnverifiable)
	require.Empty(t, payload2)
}
