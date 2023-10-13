// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/charbonnierg/caddy-nats/oauthproxy"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func init() {
	caddy.RegisterModule(OAuth2ProxyAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type OAuth2ProxyAuthCallout struct {
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
	oauthApp, err := oauthproxy.LoadApp(app.Context())
	if err != nil {
		return err
	}
	endpoint := oauthApp.GetEndpoint(c.Endpoint)
	if endpoint == nil {
		return fmt.Errorf("oauth2_proxy endpoint %s not found", c.Endpoint)
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
	userClaims.Audience = request.ConnectOptions.Username
	// Password is an HTTP cookie
	cookie, err := parseCookies(request.ConnectOptions.Password)
	if err != nil {
		return nil, fmt.Errorf("unable to parse cookie: %v", err)
	}
	// Decode the cookie
	sessionState, err := c.endpoint.DecodeSessionState(cookie)
	if err != nil {
		return nil, fmt.Errorf("unable to decode session state: %v", err)
	}
	// We've got a session state, we can now set the user claims
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
