// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/charbonnierg/caddy-nats/oauthproxy"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(OAuth2ProxyAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type OAuth2ProxyAuthCallout struct {
	logger     *zap.Logger
	sk         nkeys.KeyPair
	endpoint   *oauthproxy.Endpoint
	Endpoint   string `json:"endpoint"`
	SigningKey string `json:"signing_key"`
}

func (OAuth2ProxyAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.oauth2",
		New: func() caddy.Module { return new(OAuth2ProxyAuthCallout) },
	}
}

func (c *OAuth2ProxyAuthCallout) Provision(app *modules.App) error {
	c.logger = app.Context().Logger().Named("oauth2")
	oauthApp, err := oauthproxy.LoadApp(app.Context())
	if err != nil {
		return err
	}
	endpoint, err := oauthApp.GetEndpoint(c.Endpoint)
	if err != nil {
		return err
	}
	c.endpoint = endpoint
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

func (c *OAuth2ProxyAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	resp := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	// Use the username as the issuer account.
	// We don't look at the password
	// But in a more useful module, password could be an OpenID token maybe ?
	userClaims := jwt.NewUserClaims(request.UserNkey)
	// Username is the account. It must be set as the audience of user claims
	acc := request.ConnectOptions.Username
	userClaims.Audience = acc
	// Password is an HTTP cookie
	sessionState, err := c.endpoint.DecodeSessionStateFromString(request.ConnectOptions.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to decode session state: %v", err)
	}
	// We've got a session state, we can now set the user claims
	c.logger.Info("authenticated user", zap.String("email", sessionState.Email), zap.String("account", acc))
	userClaims.Name = sessionState.Email
	// Encode using signing key
	encoded, err := userClaims.Encode(c.sk)
	if err != nil {
		return nil, err
	}
	resp.Jwt = encoded
	return resp, nil
}

var (
	_ modules.AuthCallout = (*OAuth2ProxyAuthCallout)(nil)
)
