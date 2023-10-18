// SPDX-License-Identifier: Apache-2.0

package natsapp

import "github.com/nats-io/jwt/v2"

type AuthCallout interface {
	Handle(request *AuthorizationRequest) (*jwt.UserClaims, error)
	Provision(app *App) error
}
