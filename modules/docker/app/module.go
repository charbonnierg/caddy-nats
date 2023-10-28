// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
	interfaces "github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	ctx     caddy.Context
	logger  *zap.Logger
	beyond  *beyond.Beyond
	secrets secrets.App
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "docker",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	// This will load the beyond module and register the "secrets" app within beyond module
	b, err := beyond.Register(ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	// At this point we can use the beyond module to load other apps
	// Let's load the secret app
	unm, err := b.LoadApp(secrets.NS)
	if err != nil {
		return fmt.Errorf("failed to load secrets app: %v", err)
	}
	a.secrets = unm.(secrets.App)
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Starting docker app")
	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping docker app")
	return nil
}

func (a *App) Validate() error {
	return nil
}

var (
	// Make sure app implements the beyond.App interface
	_ beyond.App = (*App)(nil)
	// Only methods exposed by the interfaces.SecretApp interface will be accessible
	// to other apps.
	_ interfaces.App = (*App)(nil)
)
