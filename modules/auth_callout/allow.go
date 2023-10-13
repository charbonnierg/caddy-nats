// SPDX-License-Identifier: Apache-2.0

package auth_callout

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func init() {
	caddy.RegisterModule(AllowAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type AllowAuthCallout struct {
	sk         nkeys.KeyPair
	SigningKey string `json:"signing_key"`
}

func (AllowAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.always_allow",
		New: func() caddy.Module { return new(AllowAuthCallout) },
	}
}

func (c *AllowAuthCallout) Provision(app *modules.App) error {
	var seed []byte
	if app.Options.Operators == nil {
		seed = []byte(app.AuthService.AuthSigningKey)
	} else {
		seed = []byte(c.SigningKey)
	}
	sk, err := nkeys.FromSeed(seed)
	if err != nil {
		return err
	}
	c.sk = sk
	return nil
}

func (a *AllowAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	resp := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	// Use the username as the issuer account.
	// We don't look at the password
	// But in a more useful module, password could be an OpenID token maybe ?
	userClaims := jwt.NewUserClaims(request.UserNkey)
	// The target account must be specified as JWT audience
	userClaims.Audience = request.ConnectOptions.Username
	// Encode using signing key
	encoded, err := userClaims.Encode(a.sk)
	if err != nil {
		return nil, err
	}
	resp.Jwt = encoded
	return resp, nil
}

var (
	_ modules.AuthCallout = (*AllowAuthCallout)(nil)
)
