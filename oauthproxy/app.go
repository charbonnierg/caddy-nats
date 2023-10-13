package oauthproxy

import "github.com/caddyserver/caddy/v2"

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	ctx       caddy.Context
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauth2",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	for _, e := range a.Endpoints {
		if err := e.Provision(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Start() error {
	for _, e := range a.Endpoints {
		if err := e.setup(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Stop() error {
	return nil
}
