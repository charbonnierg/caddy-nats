// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import (
	"context"
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
// It implements the caddyhttp.MiddlewareHandler interface.
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

// This should not be needed now that whole config object is exposed
// but it is kept to avoid changing example config files.
func (e *Endpoint) applyDefaultOpts() error {
	// Configure cookie name
	if e.opts.Cookie.Name == "" {
		e.opts.Cookie.Name = fmt.Sprintf("%s_%s", e.Name, "oauth2_proxy")
	}
	// Configure remaining options
	e.opts.SkipJwtBearerTokens = true
	e.opts.ReverseProxy = true
	e.opts.SkipProviderButton = true
	e.opts.EmailDomains = []string{"*"}

	// This is a hack to get the access token and id token to the backend
	e.opts.InjectRequestHeaders = append(e.opts.InjectRequestHeaders, options.Header{
		Name: "X-Forwarded-Access-Token",
		Values: []options.HeaderValue{
			{
				ClaimSource: &options.ClaimSource{
					Claim: "access_token",
				},
			},
		},
	},
		options.Header{
			Name: "Authorization",
			Values: []options.HeaderValue{
				{
					ClaimSource: &options.ClaimSource{
						Claim: "id_token",
					},
				},
			},
		})
	return nil
}

// Provision loads and validate the endpoint configuration.
// It is called when AddEndpoint method of the app is called and should
// not be called directly by other host modules.
func (e *Endpoint) Provision(ctx caddy.Context) error {
	// Initialize options
	if e.Options == nil {
		return fmt.Errorf("no options found for endpoint %s", e.Name)
	}
	// Save options
	e.opts = e.Options.oauth2proxyOpts()
	// Save logger
	e.logger = ctx.Logger().Named(e.Name)

	// TODO: This should not be needed now that whole config object is exposed
	// but it is kept to avoid changing example config files.
	if err := e.applyDefaultOpts(); err != nil {
		return err
	}
	// TODO: The secret is randomely generated, but it is not persisted
	if e.opts.Cookie.Secret == "" {
		secret, err := generateRandomASCIIString(32)
		if err != nil {
			return err
		}
		e.opts.Cookie.Secret = secret
	}
	// Save cipher
	cipher, err := encryption.NewCFBCipher(encryption.SecretBytes(e.opts.Cookie.Secret))
	if err != nil {
		return fmt.Errorf("error initialising cipher: %v", err)
	}
	e.cipher = cipher
	// Validate options
	if err := validation.Validate(e.opts); err != nil {
		return err
	}

	return nil
}

// setup sets up the oauth2-proxy instance for this endpoint.
// It is called when the app is started, not when the endpoint is provisioned.
func (e *Endpoint) setup() error {
	chainer := chainer{logger: e.logger}
	validator := server.NewValidator(e.opts.EmailDomains, e.opts.AuthenticatedEmailsFile)
	proxy, err := server.NewEmbeddedOauthProxy(e.opts, validator, &chainer)
	if err != nil {
		return err
	}
	e.proxy = proxy
	return nil
}

func (e *Endpoint) isReference() bool {
	return e.Options == nil
}

func (e *Endpoint) equals(other *Endpoint) bool {
	if e.Name != other.Name {
		return false
	}
	if e.Options == nil && other.Options == nil {
		return true
	}
	return e.Options.equals(other.Options)
}

// nextKey is a struct used as key to store the next handler in the request context.
// Context docs recommends to use a private struct rather than a string to avoid
// collisions with other packages.
type nextKey struct{}

// chainer is a struct that implements the http.Handler interface.
// It is called by oauth2-proxy gorilla mux when the request is authorized.
// It fetches the next handler from the request context and calls it.
// This whole thing relies on Endpoint.ServeHTTP to set the next handler in the request context
// under the nextKey{} key.
type chainer struct {
	logger *zap.Logger
}

// ServeHTTP fetches the next handler from the request context and calls it.
// It is called by oauth2-proxy when the request is authorized as the upstream handler.
func (h chainer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("serving authorized request", zap.Any("access_token", r.Header["X-Forwarded-Access-Token"]), zap.Any("id_token", r.Header["Authorization"]))
	nextRaw := r.Context().Value(nextKey{})
	if nextRaw == nil {
		h.logger.Error("next handler not found in request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	next, ok := nextRaw.(caddyhttp.Handler)
	if !ok {
		h.logger.Error("next handler is not an http.Handler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err := next.ServeHTTP(w, r)
	if err != nil {
		h.logger.Error("error serving next handler", zap.Error(err))
	}
}
