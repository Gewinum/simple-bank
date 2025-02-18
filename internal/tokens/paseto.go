package tokens

import (
	"fmt"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
	"time"
)

type PasetoManager struct {
	paseto     *paseto.V2
	privateKey string
}

func NewPasetoManager(privateKey string) (Manager, error) {
	if len(privateKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("symmetricKey must be %d bytes long", chacha20poly1305.KeySize)
	}

	return &PasetoManager{
		paseto:     paseto.NewV2(),
		privateKey: privateKey,
	}, nil
}

func (manager *PasetoManager) CreateToken(params PayloadCreationParams) (string, error) {
	payload, err := NewPayload(params)
	if err != nil {
		return "", err
	}

	return manager.paseto.Encrypt([]byte(manager.privateKey), payload, nil)
}

func (manager *PasetoManager) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := manager.paseto.Decrypt(token, []byte(manager.privateKey), payload, nil)
	if err != nil {
		return nil, err
	}

	if payload.NotBefore.After(time.Now()) {
		return nil, ErrTokenNotValidYet
	}

	if payload.ExpiredAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return payload, nil
}
