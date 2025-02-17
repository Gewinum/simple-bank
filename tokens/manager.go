package tokens

import (
	"errors"
)

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrTokenNotValidYet = errors.New("token not valid yet")
)

type Manager interface {
	CreateToken(params PayloadCreationParams) (string, error)
	VerifyToken(token string) (*Payload, error)
}
