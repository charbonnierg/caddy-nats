// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"errors"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/quara-dev/beyond/modules/nats/natsapp"
	"github.com/quara-dev/beyond/modules/oauth2"
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
	oauth    oauth2.OAuth2App
	endpoint oauth2.OAuth2Endpoint
	Endpoint string            `json:"endpoint"`
	Account  string            `json:"account,omitempty"`
	Template *natsapp.Template `json:"template,omitempty"`
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
func (c *OAuth2ProxyAuthCallout) Provision(app *natsapp.App) error {
	c.logger = app.Context().Logger().Named("oauth2")
	// Load oauth2 app
	unm, err := app.LoadApp("oauth2")
	if err != nil {
		return err
	}
	oauthApp, ok := unm.(oauth2.OAuth2App)
	if !ok {
		return errors.New("invalid oauth2 app module")
	}
	c.oauth = oauthApp
	return nil
}

func (c *OAuth2ProxyAuthCallout) loadOauth2Endpoint() error {
	if c.endpoint != nil {
		return nil
	}
	// Load oauth2 endpoint
	endpoint, err := c.oauth.GetEndpoint(c.Endpoint)
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
func (c *OAuth2ProxyAuthCallout) Handle(request *natsapp.AuthorizationRequest) (*jwt.UserClaims, error) {
	// Endpoint will be loaded on first request, then it is saved on
	// the callout instance for subsequent requests
	err := c.loadOauth2Endpoint()
	if err != nil {
		return nil, err
	}
	// Initialize user claims
	userClaims := jwt.NewUserClaims(request.Claims.UserNkey)
	// OAuth2 session state must be presented as password in connect opts (encrypted cookie string)
	sessionState, err := c.endpoint.DecodeSessionStateFromString(request.Claims.ConnectOptions.Password)
	if err != nil {
		return nil, errors.New("unable to decode session state")
	}
	// Add replacers for session state
	c.addSessionReplacerVars(request, sessionState)
	// Set target account
	if c.Account != "" {
		// The target account must be specified as JWT audience
		userClaims.Audience = request.ReplaceAll(c.Account, "")
	} else {
		// If not specified, the target account is the username
		userClaims.Audience = request.Claims.ConnectOptions.Username
	}
	if userClaims.Audience == "" {
		// If the target account is still empty, deny access
		return nil, errors.New("no target account specified")
	}
	if c.Template != nil {
		// Apply the template
		c.Template.Render(request, userClaims)
	} else {
		// If no template is specified, set the email as user name
		userClaims.Name = sessionState.Email
	}
	c.logger.Info("authenticated user", zap.String("email", sessionState.Email), zap.String("account", userClaims.Audience))
	// And that's it, return user claims
	return userClaims, nil
}

func (c *OAuth2ProxyAuthCallout) addSessionReplacerVars(request *natsapp.AuthorizationRequest, session *sessions.SessionState) {
	extractor, err := c.endpoint.GetOidcSessionClaimExtractor(session)
	if err != nil {
		c.logger.Error("unable to get oidc session claim extractor", zap.Error(err))
		return
	}
	request.AddReplacerMapper(func(key string) (any, bool) {
		oidcPrefix := "oidc.session."
		if !strings.HasPrefix(key, oidcPrefix) {
			return nil, false
		}
		claim := strings.TrimPrefix(key, oidcPrefix)
		value, ok, err := extractor.GetClaim(claim)
		if err != nil {
			c.logger.Warn("unable to extract oidc session claim", zap.String("claim", claim), zap.Error(err))
			return nil, false
		}
		if !ok {
			c.logger.Warn("oidc session claim not found", zap.String("claim", claim))
			return nil, false
		}
		stringValue, ok := value.(string)
		if !ok {
			c.logger.Warn("oidc session claim is not a string", zap.String("claim", claim))
			return nil, false
		}
		return stringValue, true
	})
}

var (
	_ natsapp.AuthCallout = (*OAuth2ProxyAuthCallout)(nil)
)
