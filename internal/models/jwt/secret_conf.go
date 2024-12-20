package jwt

import (
	"errors"
)

type favContextKey struct{}

var (
	errEmptySecret                  = errors.New("jwt secret is empty")
	Payload           favContextKey = struct{}{}
	secretKeyProvider func() []byte
)

func SetJWTSecret(secret string) error {
	if secret == "" {
		return errEmptySecret
	}

	secretKeyProvider = func() []byte {
		return []byte(secret)
	}

	return nil
}
