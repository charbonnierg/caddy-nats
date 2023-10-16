// SPDX-License-Identifier: Apache-2.0

package auth_callout

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/nats-io/jwt/v2"
)

func init() {
	caddy.RegisterModule(AllowAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type AllowAuthCallout struct{}

func (AllowAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.always_allow",
		New: func() caddy.Module { return new(AllowAuthCallout) },
	}
}

func (c *AllowAuthCallout) Provision(app *modules.App) error {
	return nil
}

func (a *AllowAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	// Use the username as the issuer account.
	// We don't look at the password
	// But in a more useful module, password could be an OpenID token maybe ?
	userClaims := jwt.NewUserClaims(request.UserNkey)
	// The target account must be specified as JWT audience
	userClaims.Audience = request.ConnectOptions.Username
	// And that's it, return the user claims
	return userClaims, nil
}

var (
	_ modules.AuthCallout = (*AllowAuthCallout)(nil)
)
