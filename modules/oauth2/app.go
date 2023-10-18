// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond"
	"github.com/charbonnierg/beyond/modules/oauth2/interfaces"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

// App is the oauth2 app module.
// It is the root module of the oauth2 Caddy module.
// It contains a list of endpoints which are provisioned and setup when the app is started.
// Each endpoint will have an oauth2-proxy instance setup when the app is started.
// Those instances can be used to perform oauth2 auth code flow by http middlewares,
// or decode and validate session cookies by other caddy modules.
// Each instance is configured with a cookie secret which is used to encrypt session cookies.
// Those secrets are generated automatically when not provided, and in such case are not exposed to other modules.
type App struct {
	mutex     *sync.Mutex
	ctx       caddy.Context
	logger    *zap.Logger
	beyond    *beyond.Beyond
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
	a.logger = ctx.Logger()
	a.mutex = &sync.Mutex{}
	// Register module against beyond module
	b, err := beyond.RegisterApp(a.ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	// Endpoints present in the config at this point are the ones that were configured in the Caddyfile/JSON config.
	// They were not added using .GetOrAddEndpoint() because this method can only be called after the app is provisioned.
	// So we need to provision them here.
	for _, e := range a.Endpoints {
		if err := e.provision(a); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the app. It implements the caddy.App interface.
// It does not start background task, but it does setup the endpoints.
func (a *App) Start() error {
	// Setup each endpoint
	for _, e := range a.Endpoints {
		// This will setup session store, oauth2-proxy instance, and upstream handler
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

var (
	_ interfaces.OAuth2Endpoint = (*Endpoint)(nil)
	_ interfaces.OAuth2App      = (*App)(nil)
	_ beyond.App                = (*App)(nil)
)
