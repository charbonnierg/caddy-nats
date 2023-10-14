// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import "github.com/caddyserver/caddy/v2"

func init() {
	caddy.RegisterModule(new(App))
}

// App is the oauth2 app module.
// It is the root module of the oauth2 Caddy module.
// It contains a list of endpoints which are provisioned and setup when the app is started.
// Each endpoint is a valid oauth2-proxy configuration.
type App struct {
	ctx       caddy.Context
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
}

// CaddyModule returns the Caddy module information.
// It implements the caddy.Module interface.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauth2",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision sets up the app when it is first loaded.
// It implements the caddy.Provisioner interface.
func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	for _, e := range a.Endpoints {
		if err := e.Provision(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the app. It implements the caddy.App interface.
// It does not start background task, but it does setup the endpoints.
func (a *App) Start() error {
	for _, e := range a.Endpoints {
		if err := e.setup(); err != nil {
			return err
		}
	}
	return nil
}

// Stop is a no-op. It implements the caddy.App interface.
func (a *App) Stop() error {
	return nil
}
