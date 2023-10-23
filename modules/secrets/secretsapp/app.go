package secretsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
	interfaces "github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	ctx          caddy.Context
	logger       *zap.Logger
	beyond       *beyond.Beyond
	defaultStore string
	stores       map[string]interfaces.Store
	StoresRaw    map[string]json.RawMessage `json:"stores,omitempty" caddy:"namespace=secrets.store inline_key=module"`
	Automations  []*Automation              `json:"automate,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	// Initialize stores map
	a.stores = make(map[string]interfaces.Store)
	// This will load the beyond module and register the "secrets" app within beyond module
	b, err := beyond.RegisterApp(a.ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	// Let's load and provision all stores
	unm, err := ctx.LoadModule(a, "StoresRaw")
	if err != nil {
		return err
	}
	for name, modIface := range unm.(map[string]interface{}) {
		mod := modIface.(interfaces.Store)
		if err := mod.Provision(a); err != nil {
			return err
		}
		a.stores[name] = mod
		if a.defaultStore == "" {
			a.defaultStore = name
		}
	}
	// Let's load and provision all automations
	for _, automation := range a.Automations {
		if err := automation.Provision(a); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Starting secrets app")
	for _, automation := range a.Automations {
		go automation.Run()
	}
	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping secrets app")
	return nil
}

var (
	// Make sure app implements the beyond.App interface
	_ beyond.App = (*App)(nil)
	// Only methods exposed by the interfaces.SecretApp interface will be accessible
	// to other apps.
	_ interfaces.SecretApp = (*App)(nil)
)
