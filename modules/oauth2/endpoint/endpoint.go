// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package endpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/encryption"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/providers/util"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/validation"
	"github.com/oauth2-proxy/oauth2-proxy/v7/server"
	"github.com/quara-dev/beyond/modules/oauth2"
	cookiestore "github.com/quara-dev/beyond/modules/oauth2/stores/cookie"
	"github.com/quara-dev/beyond/pkg/httputils"
	"github.com/quara-dev/beyond/pkg/secretutils"
	"go.uber.org/zap"
)

// Endpoint is a Caddy module that represents an oauth2-proxy endpoint.
// It implements the caddyhttp.MiddlewareHandler interface even though it's not
// used directly as an HTTP midleware. Rather, the module `http.handlers.oauth2_session`
// is used as a middleware, and it calls the endpoint ServeHTTP method..
type Endpoint struct {
	app     oauth2.App
	logger  *zap.Logger
	cipher  encryption.Cipher
	store   oauth2.Store
	opts    *options.Options
	proxy   *server.OAuthProxy
	NameRaw string          `json:"name,omitempty"`
	Options *Options        `json:"options,omitempty"`
	Store   json.RawMessage `json:"store,omitempty" caddy:"namespace=oauth2.stores inline_key=type"`
}

func (e *Endpoint) Name() string {
	return e.NameRaw
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
func (e *Endpoint) Provision(app oauth2.App) error {
	e.app = app
	// Set logger
	if e.NameRaw == "" {
		return errors.New("endpoint name cannot be empty")
	}
	e.logger = app.Logger().Named(e.Name())
	// Set options
	if e.Options == nil {
		return fmt.Errorf("no options found for endpoint %s", e.Name())
	}
	e.opts = e.Options.oauth2proxyOptions(app.GetReplacer())
	if e.opts.Cookie.Secret == "" {
		secret, err := secretutils.GenerateRandomASCIIString(32)
		if err != nil {
			return err
		}
		e.opts.Cookie.Secret = secret
	}
	// Validate options
	if err := validation.Validate(e.opts); err != nil {
		return fmt.Errorf("invalid options for endpoint %s: %v", e.Name(), err)
	}
	// Load cipher
	cipher, err := encryption.NewCFBCipher(encryption.SecretBytes(e.opts.Cookie.Secret))
	if err != nil {
		return fmt.Errorf("error initialising cipher: %v", err)
	}
	e.cipher = cipher
	// Load session store
	if e.Store == nil {

		// Use cookie store by default
		store := &cookiestore.CookieStore{}

		err := store.Provision(e.app, &e.opts.Cookie)
		if err != nil {
			return fmt.Errorf("error provisioning cookie store for endpoint %s: %v", e.Name(), err)
		}
		e.store = store
	} else {
		unm, err := e.app.Context().LoadModule(e, "Store")
		if err != nil {
			return fmt.Errorf("error loading session store for endpoint %s: %v", e.Name(), err)
		}
		store, ok := unm.(oauth2.Store)
		if !ok {
			return fmt.Errorf("invalid session store for endpoint %s", e.Name())
		}
		err = store.Provision(e.app, &e.opts.Cookie)
		if err != nil {
			return fmt.Errorf("error provisioning session store for endpoint %s: %v", e.Name(), err)
		}
		e.store = store
	}
	return nil
}

// DecodeSessionState decodes the session state from the given cookies.
// It returns an error if the cookies are invalid or if the session state
// cannot be decoded for this endpoint.
// The cookie secret used to decode session state is not exposed as a public
// attribute or method, so it is not possible to decode session state for
// an endpoint without access to the endpoint instance.
func (e *Endpoint) DecodeSessionState(cookies []*http.Cookie) (*sessions.SessionState, error) {
	e.logger.Info("decoding session state", zap.Any("cookies", cookies))
	cookie := httputils.JoinCookies(cookies, e.opts.Cookie.Name)
	req := &http.Request{}
	req.Header = http.Header{}
	req.AddCookie(cookie)
	state, err := e.store.GetStore().Load(req)
	if err != nil {
		return nil, fmt.Errorf("failed to load session state: %v", err)
	}
	return state, nil
}

// DecodeSessionStateFromString decodes the session state from the given encoded cookie string.
// It returns an error if the cookie is invalid or if the session state
// cannot be decoded for this endpoint.
// The cookie secret used to decode session state is not exposed as a public
// attribute or method, so it is not possible to decode session state for
// an endpoint without access to the endpoint instance.
func (e *Endpoint) DecodeSessionStateFromString(cookie string) (*sessions.SessionState, error) {
	cookies := httputils.ParseCookies(cookie)
	return e.DecodeSessionState(cookies)
}

func (e *Endpoint) GetOidcSessionClaimExtractor(state *sessions.SessionState) (util.ClaimExtractor, error) {
	// FIXME: What should we do if we got multiple providers for this endpoint ?
	// I guess we should first decode ID token, then check if the issuer matches
	// a specific provider issuer, then use the profile URL from the provider.
	profileURL, err := url.Parse(e.opts.Providers[0].ProfileURL)
	if err != nil {
		return nil, err
	}
	// NewClaimExtractor expect a http.Header, so we need to create one
	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", state.IDToken))
	extractor, err := util.NewClaimExtractor(context.TODO(), state.IDToken, profileURL, headers)
	if err != nil {
		return nil, err
	}
	return extractor, nil
}

// setup sets up the oauth2-proxy instance for this endpoint.
// It is called when the app is started, not when the endpoint is provisioned.
func (e *Endpoint) Setup() error {
	up := upstream{logger: e.logger}
	validator := server.NewValidator(e.opts.EmailDomains, e.opts.AuthenticatedEmailsFile)
	proxy, err := server.NewEmbeddedOauthProxy(e.opts, validator, e.store.GetStore(), &up)
	if err != nil {
		return err
	}
	up.setSessionLoader(proxy.LoadCookiedSession)
	e.proxy = proxy
	return nil
}

// empty returns true if the endpoint has no options.
func (e *Endpoint) IsReference() bool {
	return e.Options == nil
}

// equals returns true if the endpoint has the same name and options as the other endpoint.
func (e *Endpoint) Equals(other oauth2.Endpoint) bool {
	ep2, ok := other.(*Endpoint)
	if !ok {
		return false
	}
	if e.Name() != ep2.Name() {
		return false
	}
	if !e.Options.equals(ep2.Options) {
		return false
	}
	// Compare raw bytes
	if len(e.Store) != len(ep2.Store) {
		return false
	}
	for i, v := range e.Store {
		if v != ep2.Store[i] {
			return false
		}
	}
	return true
}
