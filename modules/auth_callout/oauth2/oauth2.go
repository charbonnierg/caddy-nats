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

// Provision sets up the auth callout handler.
// It is called by the auth callout caddy module when the handler is loaded from config.
// It should not be called directly by other modules.
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

	// We should decide whether to allow the user to connect to the target account
	// For now, we allow all users to connect to all accounts
	// ...

	// Encode user claims using signing key (the nkey seed associated with the issuer public key in server mode, or a signing key of target account in operator mode)
	// This will work poorly in operator mode, as a single signing key is used for all accounts.
	// But NATS auth callout ADR explicitely states that the callout service can issue JWT for DIFFERENT accounts
	// as long as they are signed by the target account.
	// Ideally, caddy modules could be used to fetch signing key from a remote service, but that's out of scope for now.
	encoded, err := userClaims.Encode(c.sk)
	if err != nil {
		return nil, err
	}
	// Set encoded user claims as JWT in response claims
	responseClaims.Jwt = encoded
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
