package natsmagic

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/charbonnierg/caddy-nats/natsissuer"
	"go.uber.org/zap"
)

// Register caddy module when file is imported
func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("nats_server", ParseNatsOptions)
}

// App is the Caddy module that handles the configuration
// and the lifecycle of the embedded NATS server
type App struct {
	// The NATS server configuration
	NATS *NatsConfig `json:"nats,omitempty"`
	// The TLS configuration
	Policies *AppPolicies `json:"policies,omitempty"`
	// Private properties
	tls      *caddytls.TLS
	logger   *zap.Logger
	ctx      caddy.Context
	automate *caddytls.AutomateLoader
	server   *NatsMagic
	// conn   *nats.Conn
	quit chan struct{}
}

// Register caddy module app
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.server",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision caddy module app
func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.logger.Info("Provisioning NATS server")
	tlsAppIface, err := ctx.App("tls")
	if err != nil {
		return fmt.Errorf("getting tls app: %v", err)
	}
	a.tls = tlsAppIface.(*caddytls.TLS)
	issuer_, err := ctx.App("nats.issuer")
	if err == nil {
		issuer := issuer_.(*natsissuer.App)
		sys, err := issuer.GetSystemAccountPublicKey()
		if err != nil {
			return fmt.Errorf("getting system account public key: %v", err)
		}
		a.NATS.Operators = []string{}
		a.NATS.Operators = append(a.NATS.Operators, issuer.Operators...)
		a.NATS.SystemAccount = sys
		a.NATS.ResolverPreload = append(a.NATS.ResolverPreload, issuer.Accounts...)
		a.NATS.ResolverPreload = append(a.NATS.ResolverPreload, issuer.SystemAccount)
	}
	subjects := a.Policies.Subjects()
	v, err := ctx.LoadModuleByID("tls.certificates.automate", caddyconfig.JSON(subjects, nil))
	if err != nil {
		return fmt.Errorf("loading tls automate module: %v", err)
	}
	a.automate = v.(*caddytls.AutomateLoader)
	if err := a.tls.Manage(*a.automate); err != nil {
		return fmt.Errorf("managing domains: %v", err)
	}
	a.server = NewServer(a.NATS)
	return nil
}

func setDefaultPolicy(policies caddytls.ConnectionPolicies) caddytls.ConnectionPolicies {
	if len(policies) == 0 {
		return append(policies, new(caddytls.ConnectionPolicy))
	}
	return policies
}

// Validate caddy module app
func (a *App) Validate() error {
	// ensure there is at least one policy, which will act as default
	if !a.NATS.NoTLS {
		a.Policies.StandardPolicies = setDefaultPolicy(a.Policies.StandardPolicies)
		if err := a.Policies.StandardPolicies.Provision(a.ctx); err != nil {
			return err
		}
	}
	if !a.NATS.Websocket.NoTLS {
		a.Policies.WebsocketPolicies = setDefaultPolicy(a.Policies.WebsocketPolicies)
		if err := a.Policies.WebsocketPolicies.Provision(a.ctx); err != nil {
			return err
		}
	}
	if !a.NATS.LeafNode.NoTLS {
		a.Policies.LeafnodePolicies = setDefaultPolicy(a.Policies.LeafnodePolicies)
		if err := a.Policies.LeafnodePolicies.Provision(a.ctx); err != nil {
			return err
		}
	}
	if !a.NATS.MQTT.NoTLS {
		a.Policies.MQTTPolicies = setDefaultPolicy(a.Policies.MQTTPolicies)
		if err := a.Policies.MQTTPolicies.Provision(a.ctx); err != nil {
			return err
		}
	}
	a.server.SetStandardTLSConfig(a.Policies.StandardPolicies.TLSConfig(a.ctx))
	a.server.SetWebsocketTLSConfig(a.Policies.WebsocketPolicies.TLSConfig(a.ctx))
	a.server.SetLeafnodeTLSConfig(a.Policies.LeafnodePolicies.TLSConfig(a.ctx))
	a.server.SetMQTTTLSConfig(a.Policies.MQTTPolicies.TLSConfig(a.ctx))
	a.server.SetLogger(a.logger)
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
