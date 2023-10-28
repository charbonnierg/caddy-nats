// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/otelcol/components"
	"github.com/quara-dev/beyond/modules/otelcol/fieldprovider"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("otelcol", parseGlobalOption)
}

var (
	BUILD_INFO = component.BuildInfo{
		Command:     "otelcol-beyond",
		Description: "Custom OpenTelemetry Collector distribution for Beyond project",
		Version:     "0.0.1-dev",
	}
)

type App struct {
	ctx       caddy.Context
	factories otelcol.Factories
	buildInfo component.BuildInfo
	provider  otelcol.ConfigProvider
	collector *otelcol.Collector
	logger    *zap.Logger
	Config    json.RawMessage `json:"config,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Logger() *zap.Logger {
	return a.logger
}

func (a *App) Context() caddy.Context {
	return a.ctx
}
func (o *App) Start() error {
	o.logger.Warn("Starting OTEL Collector")
	o.run()
	return nil
}
func (o *App) Stop() error {
	if o.collector != nil {
		o.logger.Warn("Stopping OTEL Collector")
		o.collector.Shutdown()
	}
	return nil
}

func (o *App) Validate() error {
	return nil
}

func (o *App) run() {
	go func() {
		if err := o.collector.Run(o.ctx); err != nil {
			o.logger.Error("collector run failed", zap.Error(err))
		}
	}()
}

func (o *App) Provision(ctx caddy.Context) error {
	o.ctx = ctx
	o.logger = ctx.Logger()
	factories, err := components.Components()
	if err != nil {
		return fmt.Errorf("failed to build components: %v", err)
	}
	o.factories = factories
	o.buildInfo = BUILD_INFO
	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{"field:Config"},
			Providers: map[string]confmap.Provider{
				"field": fieldprovider.NewProvider(o),
				"file":  fileprovider.New(),
				"http":  httpsprovider.New(),
				"https": httpsprovider.New(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create config provider: %v", err)
	}
	o.provider = provider
	settings := otelcol.CollectorSettings{
		BuildInfo:               o.buildInfo,
		Factories:               o.factories,
		DisableGracefulShutdown: true,
		ConfigProvider:          o.provider,
		LoggingOptions:          []zap.Option{zap.AddCaller(), zap.Development()},
	}
	col, err := otelcol.NewCollector(settings)
	if err != nil {
		return fmt.Errorf("failed to create collector: %v", err)
	}
	o.collector = col
	if err := o.collector.DryRun(o.ctx); err != nil {
		return err
	}
	return nil
}

var (
	_ beyond.App = (*App)(nil)
)
