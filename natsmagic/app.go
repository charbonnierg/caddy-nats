package natsmagic

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"go.uber.org/zap"
)

// Register caddy module when file is imported
func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("nats-server", ParseNatsOptions)
}

// App is the Caddy module that handles the configuration
// and the lifecycle of the embedded NATS server
type App struct {
	// The NATS server configuration
	Options *Options `json:"options,omitempty"`
	// The automation policy
	Automations []caddytls.AutomationPolicy `json:"automations,omitempty"`
	// The TLS configuration
	StandardPolicies caddytls.ConnectionPolicies `json:"standard_policies,omitempty"`
	// Private properties
	tls    *caddytls.TLS
	logger *zap.Logger
	ctx    caddy.Context
	server *NatsMagic
	// conn   *nats.Conn
	quit chan struct{}
}

// Register caddy module app
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision caddy module app
func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	tlsAppIface, err := ctx.App("tls")
	if err != nil {
		return fmt.Errorf("getting tls app: %v", err)
	}
	a.tls = tlsAppIface.(*caddytls.TLS)
	for _, automation := range a.Automations {
		if len(automation.IssuersRaw) == 0 && len(a.tls.Automation.Policies) > 0 {
			automation.Issuers = append(automation.Issuers, a.tls.Automation.Policies[0].Issuers...)
		}
		if err := a.tls.AddAutomationPolicy(&automation); err != nil {
			return fmt.Errorf("adding automation policy: %v", err)
		}
	}
	for _, automation := range a.Automations {
		if len(automation.SubjectsRaw) > 0 {
			a.tls.Manage(automation.SubjectsRaw)
		}
	}
	a.server = NewServer(a.Options)
	return nil
}

// Validate caddy module app
func (a *App) Validate() error {
	// ensure there is at least one policy, which will act as default
	if len(a.StandardPolicies) == 0 {
		a.StandardPolicies = append(a.StandardPolicies, new(caddytls.ConnectionPolicy))
	}
	err := a.StandardPolicies.Provision(a.ctx)
	if err != nil {
		return fmt.Errorf("setting up connection policies: %v", err)
	}
	a.server.SetLogger(a.logger)
	if len(a.Automations) > 0 {
		a.server.SetTLSConfig(a.StandardPolicies.TLSConfig(a.ctx))
	}
	return nil
}

// Start caddy module app
func (a *App) Start() error {
	err := a.server.Start()
	if err != nil {
		return fmt.Errorf("starting server: %v", err)
	}
	// conn, err := nats.Connect("", nats.InProcessServer(a.server.ns))
	// if err != nil {
	// 	return fmt.Errorf("connecting to server: %v", err)
	// }
	// a.conn = conn
	// ticker := time.NewTicker(5 * time.Second)
	// a.quit = make(chan struct{})
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			a.conn.Publish("hello", []byte("world"))
	// 		case <-a.quit:
	// 			ticker.Stop()
	// 			return
	// 		}
	// 	}
	// }()
	return nil
}

// Stop caddy module app
func (a *App) Stop() error {
	// a.quit <- struct{}{}
	return a.server.Stop()
}

// Interface guards
var (
	_ caddy.Module      = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.App         = (*App)(nil)
	_ caddy.Validator   = (*App)(nil)
)
