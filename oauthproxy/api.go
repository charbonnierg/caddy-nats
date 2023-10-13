package oauthproxy

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/encryption"
	"go.uber.org/zap"
)

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

func (a *App) GetEndpoint(name string) *Endpoint {
	for _, e := range a.Endpoints {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (a *App) AddEndpoint(e *Endpoint) error {
	for _, endpoint := range a.Endpoints {
		if endpoint.Name == e.Name {
			return fmt.Errorf("endpoint %s already exists", e.Name)
		}
	}
	if err := e.Provision(a.ctx); err != nil {
		return err
	}
	a.Endpoints = append(a.Endpoints, e)
	return nil
}

func (e *Endpoint) DecodeSessionState(cookies []*http.Cookie) (*sessions.SessionState, error) {
	e.logger.Info("decoding session state", zap.Any("cookies", cookies))
	cookie, err := joinCookies(cookies, e.opts.Cookie.Name)
	if err != nil {
		return nil, err
	}
	val, _, ok := encryption.Validate(cookie, e.opts.Cookie.Secret, e.opts.Cookie.Expire)
	if !ok {
		return nil, errors.New("cookie signature not valid")
	}
	return sessions.DecodeSessionState(val, e.cipher, true)
}
