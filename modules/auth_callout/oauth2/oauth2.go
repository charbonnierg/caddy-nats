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

// OAuth2ProxyAuthCallout is a caddy module that implements the auth callout interface.
// It is used to authenticate users using an oauth2 proxy.
// It is configured in the "nats.auth_callout.oauth2" namespace.
// It must be configured with an endpoint name, which must be defined in the oauth2 app.
// If NATS server is running in operator mode, it must also be configured with signing keys
// for the target accounts.
// In NATS server is running in server mode, there is no need to configure signing keys,
// as the signing key is configured in the parent caddy module.
// This auth callout always expects the username to be the target account, and the password
// to be the oauth2 session state (encrypted cookie string).
// It is useful to authenticate websocket users coming from applications protected
// by an oauth2 proxy.
type OAuth2ProxyAuthCallout struct {
	logger         *zap.Logger
	auth           *modules.AuthService
	endpoint       *oauthproxy.Endpoint
	keys           map[string]nkeys.KeyPair
	Endpoint       string            `json:"endpoint"`
	SigningKeysRaw map[string]string `json:"signing_keys,omitempty"`
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
	// Save auth service
	c.auth = app.AuthService
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
// It returns unsigned response claims, as the signing key is configured in the caddy module.
// However, it must sign the user claims using either the signing key configured in the caddy module
// in server mode, or the signing key of the target account in operator mode. Since only a single static
// signing key is supported in configuration for now, it is not possible to issue JWT for different accounts
// in operator mode.
func (c *OAuth2ProxyAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	// Initialize response claims
	responseClaims := jwt.NewAuthorizationResponseClaims(request.UserNkey)
	// Initialize user claims
	userClaims := jwt.NewUserClaims(request.UserNkey)
	// Target account must be present as username in connect opts
	targetAccount := request.ConnectOptions.Username
	// Set audience to target account
	userClaims.Audience = targetAccount
	// OAuth2 session state must be presented as password in connect opts (encrypted cookie string)
	sessionState, err := c.endpoint.DecodeSessionStateFromString(request.ConnectOptions.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to decode session state: %v", err)
	}
	// We've got a session state, we can now set the user claims
	c.logger.Info("authenticated user", zap.String("email", sessionState.Email), zap.String("account", targetAccount))
	userClaims.Name = sessionState.Email

	// Ideally, caddy modules could be used to fetch signing key from a remote service, but that's out of scope for now.
	if c.keys != nil {
		// We are in operator mode, use the signing key of the target account
		key, ok := c.keys[targetAccount]
		if !ok {
			return nil, fmt.Errorf("no signing key for account %s", targetAccount)
		}
		encoded, err := userClaims.Encode(key)
		if err != nil {
			return nil, err
		}
		// Set encoded user claims as JWT in response claims
		responseClaims.Jwt = encoded
	} else {
		// We are in server mode, use the signing key configured in the caddy module
		encoded, err := c.auth.SignUserClaims(userClaims)
		if err != nil {
			return nil, err
		}
		// Set encoded user claims as JWT in response claims
		responseClaims.Jwt = encoded
	}
	// Return response claims (no need to sign them, as the caddy module will do it for us)
	// In operator mode, this must be the signing key of the auth account in which auth callout is configured.
	// In server mode, this is the nkey seed associated with the issuer public key.
	// In both cases, a single signing key is used for all responses, because:
	//   - in server mode there is a single issuer, it makes sense to configure a single signing key.
	//   - in operator mode, a service can only connect to a single account, it also makes sense to configure a single signing key.
	return responseClaims, nil
}

var (
	_ modules.AuthCallout = (*OAuth2ProxyAuthCallout)(nil)
)
