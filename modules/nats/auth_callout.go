package nats

import "github.com/nats-io/jwt/v2"

type AuthCallout interface {
	Handle(request *AuthorizationRequest) (*jwt.UserClaims, error)
	Provision(app *App) error
}
