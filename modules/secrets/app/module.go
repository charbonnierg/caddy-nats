// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

// Register the secrets app on module load
// and register "secrets" caddyfile global option
func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption(secrets.NS, parseGlobalOption)
}

// App is the secrets caddy app module. It defines the "secrets" namespace.
// It can be provisioned with stores and automations.
// Stores are used to retrieve secrets from various sources.
// Automations are used to periodically retrieve secrets from stores and
// execute actions with the retrieved secrets.
// This module is registered as a caddy app and as a beyond app.
type App struct {
	StoresRaw      map[string]json.RawMessage `json:"stores,omitempty" caddy:"namespace=secrets.stores inline_key=module"`
	AutomationsRaw []json.RawMessage          `json:"automate,omitempty" caddy:"namespace=secrets.automation inline_key=type"`

	ctx          caddy.Context
	logger       *zap.Logger
	automations  []secrets.Automation
	stores       secrets.Stores
	defaultStore string
}

// CaddyModule returns the Caddy module information.
// This method is required to implement the beyond.App interface
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
	_, err := beyond.Register(ctx, a)
	if err != nil {
		return err
	}
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
// This method is required to implement the beyond.App interface
func (a *App) Start() error {
	a.logger.Info("Starting secrets app")
	for _, automation := range a.automations {
		automation.Start()
	}
	return nil
}

// Stop stops the secrets app.
// This method is required to implement the beyond.App interface
func (a *App) Stop() error {
	a.logger.Info("Stopping secrets app")
	for _, automation := range a.automations {
		automation.Stop()
	}
	return nil
}

// Validate checks if the secrets app is valid.
// This method is required to implement the beyond.App interface
func (a *App) Validate() error {
	if a.defaultStore != "" {
		_, ok := a.stores[a.defaultStore]
		if !ok {
			return secrets.ErrStoreNotFound
		}
	}
	return nil
}

// loadStoresRaw is used to load StoresRaw property into stores property
func (a *App) loadStoresRaw() error {
	// Let's load and provision all stores
	unm, err := a.ctx.LoadModule(a, "StoresRaw")
	if err != nil {
		return err
	}
	for name, modIface := range unm.(map[string]interface{}) {
		store, ok := modIface.(secrets.Store)
		if !ok {
			return errors.New("failed to load store")
		}
		if err := a.AddStore(name, store); err != nil {
			return err
		}
	}
	return nil
}

// loadAutomations is used to provision all automations found in Automations property
func (a *App) loadAutomations() error {
	// Let's load and provision all automations
	for _, raw := range a.AutomationsRaw {
		unm, err := a.ctx.LoadModuleByID("secrets.automation", raw)
		if err != nil {
			return err
		}
		automation, ok := unm.(secrets.Automation)
		if !ok {
			return errors.New("failed to load automation")
		}
		if err := a.AddAutomation(automation); err != nil {
			return err
		}
	}
	return nil
}

// Interface guards
var (
	// Make sure app implements the beyond.App interface
	_ beyond.App = (*App)(nil)
	// Only methods exposed by the secrets.SecretApp interface will be accessible
	// to other apps.
	_ secrets.App = (*App)(nil)
)
