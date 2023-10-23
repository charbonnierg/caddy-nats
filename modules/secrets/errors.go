package secrets

import "errors"

var (
	ErrSecretNotFound      = errors.New("secret not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
)
