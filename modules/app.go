// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/charbonnierg/caddy-nats/embedded/natsoptions"
	"github.com/charbonnierg/caddy-nats/embedded/natsrunner"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	ctx                caddy.Context
	tlsApp             *caddytls.TLS
	logger             *zap.Logger
	runner             *natsrunner.Runner
	connectionPolicies []caddytls.ConnectionPolicies
	subjects           []string
	AuthService        *AuthService         `json:"auth_service,omitempty"`
	Options            *natsoptions.Options `json:"server,omitempty"`
	ReadyTimeout       time.Duration        `json:"ready_timeout,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	var err error
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.logger.Info("Provisioning NATS server")
	// Make sure options exist
	if a.Options == nil {
		a.Options = &natsoptions.Options{}
	}
	// Provision tls app and connections policies
	a.connectionPolicies = []caddytls.ConnectionPolicies{}
	tlsunm, err := ctx.App("tls")
	if err != nil {
		return errors.New("failed to get tls app")
	}
	tlsApp, ok := tlsunm.(*caddytls.TLS)
	if !ok {
		return errors.New("tls app invalid type")
	}
	a.tlsApp = tlsApp
	// Provision auth service
	if a.AuthService != nil {
		if err := a.AuthService.Provision(a); err != nil {
			return err
		}
	}
	// We could update options here if we want
	// For example we could set the TLS config override
	// of the standard NATS server to use ACME certificates:
	if err := a.setTLSConfigOverride(); err != nil {
		return err
	}
	// Create runner
	a.runner, err = natsrunner.New().
		WithOptions(a.Options).
		WithLogger(a.logger.Named("server")).
		WithReadyTimeout(a.ReadyTimeout).
		Build()
	// Fail if runner creation failed
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Managing TLS certificates", zap.Strings("subjects", a.subjects))
	if a.subjects != nil {
		if err := a.tlsApp.Manage(a.subjects); err != nil {
			return err
		}
	}
	// Start nats runner
	if err := a.runner.Start(); err != nil {
		return err
	}
	// Start auth service
	if a.AuthService != nil {
		if err := a.AuthService.Start(a.runner.Server()); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Stop() error {
	// Stop auth service
	if a.AuthService != nil {
		if err := a.AuthService.Stop(); err != nil {
			a.logger.Error("Failed to stop auth service", zap.Error(err))
		}
	}
	// Stop nats runner
	return a.runner.Stop()
}

func (a *App) setStandardTLSConnectionPolicies() caddytls.ConnectionPolicies {
	if a.Options.TLS == nil || a.Options.TLS.Subjects == nil {
		return nil
	}
	subjects := a.Options.TLS.Subjects
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
	if a.Options.Websocket == nil || a.Options.Websocket.TLS == nil || a.Options.Websocket.TLS.Subjects == nil {
		return nil
	}
	subjects := a.Options.Websocket.TLS.Subjects
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
	if a.Options.Leafnode == nil || a.Options.Leafnode.TLS == nil || a.Options.Leafnode.TLS.Subjects == nil {
		return nil
	}
	subjects := a.Options.Leafnode.TLS.Subjects
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
	if a.Options.TLS == nil {
		return nil
	}
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
		a.Options.TLS.SetConfigOverride(standardPolicies.TLSConfig(a.ctx))
	}
	if wsPolicies != nil {
		a.logger.Debug("Setting Websocket TLS config override", zap.Any("policies", wsPolicies))
		tlsConfig := wsPolicies.TLSConfig(a.ctx)
		a.Options.Websocket.TLS.SetConfigOverride(tlsConfig)
	}
	if leafPolicies != nil {
		a.logger.Debug("Setting Leafnode TLS config override", zap.Any("policies", leafPolicies))
		a.Options.Leafnode.TLS.SetConfigOverride(leafPolicies.TLSConfig(a.ctx))
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
	_ caddy.App         = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
)
