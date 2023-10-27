// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/modules/secrets/automation"
	"go.uber.org/zap"
)

// Register the secrets app on module load
// and register "secrets" caddyfile global option
func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("secrets", parseGlobalOption)
}

// App is the secrets caddy app module. It defines the "secrets" namespace.
// It can be provisioned with stores and automations.
// Stores are used to retrieve secrets from various sources.
// Automations are used to periodically retrieve secrets from stores and
// execute actions with the retrieved secrets.
// This module is registered as a caddy app and as a beyond app.
type App struct {
	beyond       *beyond.Beyond
	ctx          caddy.Context
	logger       *zap.Logger
	stores       secrets.Stores
	defaultStore string
	StoresRaw    map[string]json.RawMessage `json:"stores,omitempty" caddy:"namespace=secrets.stores inline_key=module"`
	Automations  []*automation.Automation   `json:"automate,omitempty"`
}

// Private method to load StoresRaw property into stores property
func (a *App) loadStoresRaw() error {
	// Let's load and provision all stores
	unm, err := a.ctx.LoadModule(a, "StoresRaw")
	if err != nil {
		return err
	}
	for name, modIface := range unm.(map[string]interface{}) {
		mod := modIface.(secrets.Store)
		if err := mod.Provision(a); err != nil {
			return err
		}
		a.stores[name] = mod
		// Set default store if it is not already set
		if a.defaultStore == "" {
			a.defaultStore = name
		}
	}
	return nil
}

// Private method to load Automations property
func (a *App) loadAutomations() error {
	// Let's load and provision all automations
	for _, automation := range a.Automations {
		if err := automation.Provision(a); err != nil {
			return err
		}
	}
	return nil
}

// CaddyModule returns the Caddy module information.
// This method is required to implement the caddy.App interface.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision sets up the secrets app on configuration load.
// This method is required to implement the caddy.Provisioner interface.
func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.stores = secrets.Stores{}
	// This will load the beyond module and register the "secrets" app within beyond module
	b, err := beyond.RegisterApp(a.ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	// Let's load and provision all stores
	if err := a.loadStoresRaw(); err != nil {
		return err
	}
	// Let's load and provision all automations
	if err := a.loadAutomations(); err != nil {
		return err
	}
	return nil
}

// Start starts the secrets app.
// This method is required to implement the caddy.App interface.
func (a *App) Start() error {
	a.logger.Info("Starting secrets app")
	for _, automation := range a.Automations {
		go automation.Run()
	}
	return nil
}

// Stop stops the secrets app.
// This method is required to implement the caddy.App interface.
func (a *App) Stop() error {
	a.logger.Info("Stopping secrets app")
	return nil
}

// Interface guards
var (
	// Make sure app implements the beyond.App interface
	_ beyond.App = (*App)(nil)
	// Only methods exposed by the secrets.SecretApp interface will be accessible
	// to other apps.
	_ secrets.SecretApp = (*App)(nil)
)
