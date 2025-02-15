package security

import (
	"errors"
	"github.com/alexedwards/argon2id"
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
		return errors.New("password does not match")
	}
	return nil
}
