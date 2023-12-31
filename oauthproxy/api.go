// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/providers/util"
	"go.uber.org/zap"
)

// LoadApp loads the oauth2 app module from the provided caddy context.
// It is a helper function to avoid having to type assert the app module
// every time it is used, or to wrap the error in a fmt.Errorf.
func LoadApp(ctx caddy.Context) (*App, error) {
	unm, err := ctx.App("oauth2")
	if err != nil {
		return nil, fmt.Errorf("unable to get oauth2 app: %v", err)
	}
	app, ok := unm.(*App)
	if !ok {
		return nil, fmt.Errorf("invalid oauth2 app module")
	}
	return app, nil
}

// GetLogger returns a child logger with the given name.
// The logger is created using the app logger.
func (a *App) GetLogger(name string) *zap.Logger {
	return a.logger.Named(name)
}

// GetOrAddEndpoint returns the endpoint with the given name, or adds it to the app if it does not exist.
// If the endpoint already exists with different configuration, an error is returned.
// If the endpoint is added to the app, it is provisioned before being returned.
// An error is returned when the endpoint does not exist yet and cannot be provisioned.
func (a *App) GetOrAddEndpoint(e *Endpoint) (*Endpoint, error) {
	for _, existing := range a.Endpoints {
		if e.Name == existing.Name {
			if e.empty() {
				return existing, nil
			}
			if !existing.equals(e) {
				return nil, fmt.Errorf("endpoint %s already exists with different configuration", e.Name)
			}
			return existing, nil
		}
	}
	return e, a.AddEndpoint(e)
}

// GetEndpoint returns the endpoint with the given name, or an error if it does not exist.
// Error message is prefixed with "unknown oauth2 endpoint: ".
// Endpoints must be provisioned using AddEnpoint method before they can be retrieved.
func (a *App) GetEndpoint(name string) (*Endpoint, error) {
	for _, e := range a.Endpoints {
		if e.Name == name {
			return e, nil
		}
	}
	return nil, fmt.Errorf("unknown oauth2 endpoint: %s", name)
}

// AddEndpoint adds the given endpoint to the app.
// If an endpoint with the same name already exists, an error is returned.
// Otherwise, the endpoint is provisioned and added to the app.
// It can later be retrieved using GetEndpoint method.
func (a *App) AddEndpoint(e *Endpoint) error {
	for _, endpoint := range a.Endpoints {
		if endpoint.Name == e.Name {
			return fmt.Errorf("endpoint %s already exists", e.Name)
		}
	}
	if err := e.provision(a); err != nil {
		return err
	}
	a.Endpoints = append(a.Endpoints, e)
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
	cookie, err := joinCookies(cookies, e.opts.Cookie.Name)
	if err != nil {
		return nil, err
	}
	req := &http.Request{}
	req.Header = http.Header{}
	req.AddCookie(cookie)
	state, err := e.store.Store().Load(req)
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
	cookies, err := parseCookies(cookie)
	if err != nil {
		return nil, err
	}
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
