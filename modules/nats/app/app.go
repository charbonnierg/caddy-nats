// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/auth"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("nats", parseGlobalOption)
	httpcaddyfile.RegisterGlobalOption("broker", parseGlobalOption)
}

// App is the nats app module.
// It is the root module of the nats Caddy module.
// It may define options for a nats server, in which case it will start a nats server.
// It may also define an auth service, in which case it will start an auth service (as described in NATS ADR-26)
type App struct {
	Config
	ctx                caddy.Context
	beyond             *beyond.Beyond
	secrets            secrets.App
	tlsApp             *caddytls.TLS
	logger             *zap.Logger
	runner             *embedded.Runner
	connectors         []nats.Connector
	connectionPolicies []caddytls.ConnectionPolicies
	subjects           []string
	options            *embedded.Options
	authService        nats.AuthService
}

// CaddyModule returns the Caddy module information.
// It is required to implement the beyond.App interface.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats",
		New: func() caddy.Module { return new(App) },
	}
}

// Provision sets up the app when it is first loaded.
// It validates and sets up a nats server if options are defined.
// It also provisions caddy TLS connection policies for the nats server when needed,
// in order to generate TLS configs for the nats server.
// It is required to implement the beyond.App interface.
func (a *App) Provision(ctx caddy.Context) error {
	var err error
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.logger.Info("Provisioning NATS server")
	a.connectors = []nats.Connector{}
	// Provision auth service
	// This may require the Oauth2 app to be registered.
	// But the Oauth2 module itself may depend on the NATS module.
	// As such, we must register the NATS module before provisioning auth service.
	// Register module against beyond module
	b, err := beyond.Register(a.ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	// Create or load server options
	if a.ServerRaw != nil {
		options := *a.ServerRaw
		a.options = &options
	}
	// Load secrets app module
	unm, err := a.beyond.LoadApp(secrets.NS)
	if err != nil {
		return err
	}
	secretsapp, ok := unm.(secrets.App)
	if !ok {
		return errors.New("secrets app invalid type")
	}
	a.secrets = secretsapp
	// Provision tls app and connections policies
	a.connectionPolicies = []caddytls.ConnectionPolicies{}
	tlsApp, err := a.beyond.LoadTLSApp()
	if err != nil {
		return errors.New("tls app invalid type")
	}
	a.tlsApp = tlsApp
	// Now we can provision the auth service
	if a.AuthServiceRaw != nil {
		a.authService = &auth.AuthService{
			AuthServiceConfig: *a.AuthServiceRaw,
		}
		if err := a.authService.Provision(a); err != nil {
			return err
		}
	}
	// And provision the connectors
	if a.ConnectorsRaw != nil {
		unm, err := ctx.LoadModule(a, "ConnectorsRaw")
		if err != nil {
			return err
		}
		for _, v := range unm.([]interface{}) {
			connector, ok := v.(nats.Connector)
			if !ok {
				return errors.New("invalid connector type")
			}
			if err := connector.Provision(a); err != nil {
				return err
			}
			a.connectors = append(a.connectors, connector)
		}
	}
	// Replace secrets replacer variables
	replacer := caddy.NewReplacer()
	a.secrets.AddSecretsReplacerVars(replacer)
	name, err := replacer.ReplaceOrErr(a.options.ServerName, true, true)
	if err != nil {
		return fmt.Errorf("invalid secret placeholder in server name: %v", err)
	}
	a.options.ServerName = name
	a.logger.Warn("configuring nats server", zap.String("server_name", a.options.ServerName))
	// Set the TLS config override of the standard NATS server to use ACME certificates
	if err := a.setTLSConfigOverride(); err != nil {
		return err
	}
	a.logger.Debug("NATS server options", zap.Any("options", a.options))
	// Create runner
	a.runner, err = embedded.New().
		WithOptions(a.options).
		WithLogger(a.logger.Named("server")).
		WithReadyTimeout(a.ReadyTimeout).
		Build()
	// Fail if runner creation failed
	if err != nil {
		return fmt.Errorf("failed to create nats server runner: %w", err)
	}
	return nil
}

// Start starts the app.
// It starts the nats server and the auth service if defined.
// It is required to implement the beyond.App interface.
func (a *App) Start() error {
	if a.subjects != nil {
		a.logger.Info("Managing TLS certificates", zap.Strings("subjects", a.subjects))
		if err := a.tlsApp.Manage(a.subjects); err != nil {
			return err
		}
	}
	// Start nats runner
	if err := a.runner.Start(); err != nil {
		return err
	}
	// Start auth service
	if a.authService != nil {
		if err := a.startAuthService(); err != nil {
			return err
		}
	}
	// Start connectors
	for _, connector := range a.connectors {
		if err := connector.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) startAuthService() error {
	if a.authService == nil {
		return nil
	}
	// Gather auth service config and client
	cfg := a.authService.Config()
	client := a.authService.Client()
	// Set internal provider if needed
	if client.Internal {
		client.SetInternalProvider(a)
	}
	// Set username and password if needed
	u, p, err := a.GetAuthUserPass()
	if err == nil {
		client.Username = u
		client.Password = p
	}
	// Create service
	factory, err := natsutils.NewAuthServiceFactory(cfg)
	if err != nil {
		return err
	}
	// Add service to client
	client = client.AddService(factory.NewService())
	// Open connection
	_, err = client.Connect()
	if err != nil {
		return err
	}
	return nil
}

// Stop stops the app.
// It stops the nats server and the auth service if defined.
// It is required to implement the beyond.App interface.
func (a *App) Stop() error {
	// Stop connectors
	for _, connector := range a.connectors {
		if err := connector.Stop(); err != nil {
			a.logger.Error("Failed to stop connector", zap.Error(err))
		}
	}
	// Stop auth service
	if a.authService != nil {
		if client := a.authService.Client(); client != nil {
			client.Close()
		}
	}
	// Stop nats runner
	if a.runner != nil {
		a.logger.Info("Stopping NATS server")
		return a.runner.Stop()
	}
	return nil
}

// Validate is a no-op.
// It is required to implement the beyond.App interface.
func (a *App) Validate() error {
	return nil
}

func (a *App) setStandardTLSConnectionPolicies() caddytls.ConnectionPolicies {
	if a.options.TLS == nil || a.options.TLS.Subjects == nil {
		return nil
	}
	subjects := a.options.TLS.Subjects
	matcher := caddyconfig.JSON(subjects, nil)
	policy := caddytls.ConnectionPolicy{
		MatchersRaw: map[string]json.RawMessage{
			"sni": matcher,
		},
	}
	policies := caddytls.ConnectionPolicies{&policy}
	a.connectionPolicies = append(a.connectionPolicies, policies)
	return policies
}

func (a *App) setWebsocketTLSConnectionPolicies() caddytls.ConnectionPolicies {
	if a.options.Websocket == nil || a.options.Websocket.TLS == nil || a.options.Websocket.TLS.Subjects == nil {
		return nil
	}
	subjects := a.options.Websocket.TLS.Subjects
	matcher := caddyconfig.JSON(subjects, nil)
	policy := caddytls.ConnectionPolicy{
		MatchersRaw: map[string]json.RawMessage{
			"sni": matcher,
		},
	}
	policies := caddytls.ConnectionPolicies{&policy}
	a.connectionPolicies = append(a.connectionPolicies, policies)
	return policies
}

func (a *App) setLeafnodeTLSConnectionPolicies() caddytls.ConnectionPolicies {
	if a.options.Leafnode == nil || a.options.Leafnode.TLS == nil || a.options.Leafnode.TLS.Subjects == nil {
		return nil
	}
	subjects := a.options.Leafnode.TLS.Subjects
	matcher := caddyconfig.JSON(subjects, nil)
	policy := caddytls.ConnectionPolicy{
		MatchersRaw: map[string]json.RawMessage{
			"sni": matcher,
		},
	}
	policies := caddytls.ConnectionPolicies{&policy}
	a.connectionPolicies = append(a.connectionPolicies, policies)
	return policies
}

func (a *App) setTLSConfigOverride() error {
	// Set standard TLS connection policies
	standardPolicies := a.setStandardTLSConnectionPolicies()
	wsPolicies := a.setWebsocketTLSConnectionPolicies()
	leafPolicies := a.setLeafnodeTLSConnectionPolicies()
	// Gather all subjects
	subjects, err := a.findAllSubjects()
	if err != nil {
		return err
	}
	a.subjects = subjects
	// Provision connection policies
	for _, policies := range a.connectionPolicies {
		if err := policies.Provision(a.ctx); err != nil {
			return err
		}
	}
	// Now that we have the connection policies, we can set the TLS config override
	if standardPolicies != nil {
		a.logger.Debug("Setting TLS config override", zap.Any("policies", standardPolicies))
		a.options.TLS.SetConfigOverride(standardPolicies.TLSConfig(a.ctx))
	}
	if wsPolicies != nil {
		a.logger.Debug("Setting Websocket TLS config override", zap.Any("policies", wsPolicies))
		a.options.Websocket.TLS.SetConfigOverride(wsPolicies.TLSConfig(a.ctx))
	}
	if leafPolicies != nil {
		a.logger.Debug("Setting Leafnode TLS config override", zap.Any("policies", leafPolicies))
		a.options.Leafnode.TLS.SetConfigOverride(leafPolicies.TLSConfig(a.ctx))
	}
	return nil
}

func (a *App) findAllSubjects() ([]string, error) {
	a.logger.Debug("All connection policies", zap.Any("policies", a.connectionPolicies))
	subjectSet := map[string]struct{}{}
	for _, policies := range a.connectionPolicies {
		for _, pol := range policies {
			unm, err := a.ctx.LoadModule(pol, "MatchersRaw")
			if err != nil {
				return nil, err
			}
			for mod, v := range unm.(map[string]interface{}) {
				if mod != "sni" {
					continue
				}
				matcher, ok := v.(*caddytls.MatchServerName)
				if !ok {
					return nil, errors.New("internal server error: invalid matcher type")
				}
				for _, s := range *matcher {
					subjectSet[s] = struct{}{}
				}
			}
		}
	}
	subjects := make([]string, 0, len(subjectSet))
	for s := range subjectSet {
		subjects = append(subjects, s)
	}
	if len(subjects) == 0 {
		return nil, nil
	}
	return subjects, nil
}

var (
	_ nats.App = (*App)(nil)
)
