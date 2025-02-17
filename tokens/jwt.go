package tokens

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

const (
	minSecretKeyLength = 32
)

type JWTManager struct {
	secretKey string
}

func NewJWTManager(secretKey string) (Manager, error) {
	if len(secretKey) < minSecretKeyLength {
		return nil, fmt.Errorf("invalid key size: it should be at least %d characters", minSecretKeyLength)
	}

	return &JWTManager{
		secretKey: secretKey,
	}, nil
}

func (m *JWTManager) CreateToken(params PayloadCreationParams) (string, error) {
	payload, err := NewPayload(params)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) VerifyToken(token string) (*Payload, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(m.secretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrTokenInvalid
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		} else {
			return nil, err
		}
	}

	claims, ok := jwtToken.Claims.(*Payload)
	if !ok || !jwtToken.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
