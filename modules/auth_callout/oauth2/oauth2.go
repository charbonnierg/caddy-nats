// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/charbonnierg/caddy-nats/oauthproxy"
	"github.com/nats-io/jwt/v2"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(OAuth2ProxyAuthCallout{})
}

// OAuth2ProxyAuthCallout is a caddy module that implements the auth callout interface.
// It is used to authenticate users using an oauth2 proxy.
// It is configured in the "nats.auth_callout.oauth2" namespace.
// It must be configured with an endpoint name, which must be defined in the oauth2 app.
// This auth callout always expects the username to be the target account, and the password
// to be the oauth2 session state (encrypted cookie string).
type OAuth2ProxyAuthCallout struct {
	logger   *zap.Logger
	endpoint *oauthproxy.Endpoint
	Endpoint string `json:"endpoint"`
}

func (OAuth2ProxyAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.oauth2",
		New: func() caddy.Module { return new(OAuth2ProxyAuthCallout) },
	}
}

// Provision sets up the auth callout handler.
// It is called by the auth callout caddy module when the handler is loaded from config.
// It should not be called directly by other modules.
func (c *OAuth2ProxyAuthCallout) Provision(app *modules.App) error {
	c.logger = app.Context().Logger().Named("oauth2")
	// Load oauth2 app
	oauthApp, err := oauthproxy.LoadApp(app.Context())
	if err != nil {
		return err
	}
	// Load oauth2 endpoint
	endpoint, err := oauthApp.GetEndpoint(c.Endpoint)
	if err != nil {
		return err
	}
	c.endpoint = endpoint
	return nil
}

// Handle is called by auth callout caddy module to authenticate a user.
// It returns either user claims or an error.
// The account for which the user is authenticated is the username in connect opts.
// This target account is set as Audience in the user claims as required auth_callout caddy module.
func (c *OAuth2ProxyAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	// Initialize user claims
	userClaims := jwt.NewUserClaims(request.UserNkey)
	// Target account must be present as username in connect opts
	targetAccount := request.ConnectOptions.Username
	// Set audience to target account
	userClaims.Audience = targetAccount
	// OAuth2 session state must be presented as password in connect opts (encrypted cookie string)
	sessionState, err := c.endpoint.DecodeSessionStateFromString(request.ConnectOptions.Password)
	if err != nil {
		return nil, errors.New("unable to decode session state")
	}
	// We've got a session state, we can now set the user claims
	c.logger.Info("authenticated user", zap.String("email", sessionState.Email), zap.String("account", targetAccount))
	userClaims.Name = sessionState.Email
	// And that's it, return user claims
	return userClaims, nil
}

var (
	_ modules.AuthCallout = (*OAuth2ProxyAuthCallout)(nil)
)
