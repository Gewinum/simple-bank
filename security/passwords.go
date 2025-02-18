package security

import (
	"errors"
	"github.com/alexedwards/argon2id"
)

var (
	ErrPasswordNotMatched = errors.New("password does not match")
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func ComparePasswordAndHash(password, hash string) error {
	matches, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return err
	}
	if !matches {
		return ErrPasswordNotMatched
	}
	return nil
}
