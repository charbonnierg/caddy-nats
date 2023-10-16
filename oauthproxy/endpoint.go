// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/encryption"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/validation"
	"github.com/oauth2-proxy/oauth2-proxy/v7/server"
	"go.uber.org/zap"
)

// Endpoint is a Caddy module that represents an oauth2-proxy endpoint.
// It implements the caddyhttp.MiddlewareHandler interface even though it's not
// used directly as an HTTP midleware. Rather, the module `http.handlers.oauth2_session`
// is used as a middleware, and it calls the endpoint ServeHTTP method..
type Endpoint struct {
	logger  *zap.Logger
	cipher  encryption.Cipher
	opts    *options.Options
	proxy   *server.OAuthProxy
	Name    string   `json:"name,omitempty"`
	Options *Options `json:"options,omitempty"`
}

// CaddyModule returns the Caddy module information.
// It implements the caddy.Module interface.
func (Endpoint) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauthproxy.endpoint",
		New: func() caddy.Module { return new(Endpoint) },
	}
}

// ServeHTTP sets the next handler in the request context and calls the oauth2-proxy handler.
// It implements the caddyhttp.MiddlewareHandler interface.
// It is responsible for saving the next handler in the request context and calling the oauth2-proxy gorilla
// mux handler: https://github.com/oauth2-proxy/oauth2-proxy/blob/131d0b1fd2aeaf7d3456ff094ade62e448a60cf0/server/oauthproxy.go#L319
// The goal here is to call the oauth2-proxy handle and THEN the next handler.
// But because oauth2-proxy is not a caddy module, we cannot use the caddy middleware chaining mechanism.
// Instead, we configure oauth2-proxy (in the Endpoint.setup method) to use a custom upstream handler that will fetch the next handler
// from the request context and call it. Checkout the oauth2-proxy code which uses this handler:
// https://github.com/oauth2-proxy/oauth2-proxy/blob/131d0b1fd2aeaf7d3456ff094ade62e448a60cf0/server/oauthproxy.go#L984
// The upstream handler is called ONLY when the request is authorized.
func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	r = r.WithContext(context.WithValue(r.Context(), nextKey{}, next))
	e.proxy.ServeHTTP(w, r)
	return nil
}

// provision loads and validate the endpoint configuration.
// It is called when AddEndpoint method of the app is called and should
// not be called directly by other host modules.
func (e *Endpoint) provision(app *App) error {
	// Set logger
	if e.Name == "" {
		return errors.New("endpoint name cannot be empty")
	}
	e.logger = app.GetLogger(e.Name)
	// Set options
	if e.Options == nil {
		return fmt.Errorf("no options found for endpoint %s", e.Name)
	}
	e.opts = e.Options.oauth2proxyOptions()
	// Set cookie secret
	// TODO: The secret is randomely generated, but it is not persisted
	if e.opts.Cookie.Secret == "" {
		secret, err := generateRandomASCIIString(32)
		if err != nil {
			return err
		}
		e.opts.Cookie.Secret = secret
	}
	// Validate options
	if err := validation.Validate(e.opts); err != nil {
		return fmt.Errorf("invalid options for endpoint %s: %v", e.Name, err)
	}
	// Load cipher
	cipher, err := encryption.NewCFBCipher(encryption.SecretBytes(e.opts.Cookie.Secret))
	if err != nil {
		return fmt.Errorf("error initialising cipher: %v", err)
	}
	e.cipher = cipher

	return nil
}

// setup sets up the oauth2-proxy instance for this endpoint.
// It is called when the app is started, not when the endpoint is provisioned.
func (e *Endpoint) setup() error {
	up := upstream{logger: e.logger}
	validator := server.NewValidator(e.opts.EmailDomains, e.opts.AuthenticatedEmailsFile)
	proxy, err := server.NewEmbeddedOauthProxy(e.opts, validator, &up)
	up.setSessionLoader(proxy.LoadCookiedSession)
	if err != nil {
		return err
	}
	e.proxy = proxy
	return nil
}

// empty returns true if the endpoint has no options.
func (e *Endpoint) empty() bool {
	return e.Options == nil
}

// equals returns true if the endpoint has the same name and options as the other endpoint.
func (e *Endpoint) equals(other *Endpoint) bool {
	if e.Name != other.Name {
		return false
	}
	if e.Options == nil && other.Options == nil {
		return true
	}
	return e.Options.equals(other.Options)
}
