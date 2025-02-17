package tokens

import (
	"github.com/stretchr/testify/require"
	"simple-bank/random"
	"testing"
	"time"
)

func TestPasetoManager_Success(t *testing.T) {
	manager, err := NewPasetoManager(random.String(32))
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

func TestPasetoManager_ExpiredToken(t *testing.T) {
	manager, err := NewPasetoManager(random.String(32))
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
	require.EqualError(t, err, ErrTokenExpired.Error())
	require.Empty(t, payload)
}

func TestPasetoManager_SignatureInvalid(t *testing.T) {
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
	require.EqualError(t, err, ErrTokenInvalid.Error())
	require.Empty(t, payload)
}

func TestPasetoManager_UsingTokenEarly(t *testing.T) {
	manager, err := NewPasetoManager(random.String(32))
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
