// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package otelcol

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/otelcol/components"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(OtelCollector))
}

type OtelCollector struct {
	ctx       caddy.Context
	factories otelcol.Factories
	buildInfo component.BuildInfo
	provider  otelcol.ConfigProvider
	collector *otelcol.Collector
	logger    *zap.Logger
	Config    json.RawMessage `json:"config,omitempty"`
}

func (OtelCollector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol",
		New: func() caddy.Module { return new(OtelCollector) },
	}
}

func (o *OtelCollector) Start() error {
	o.logger.Warn("Starting OTEL Collector")
	o.run()
	return nil
}
func (o *OtelCollector) Stop() error {
	if o.collector != nil {
		o.logger.Warn("Stopping OTEL Collector")
		o.collector.Shutdown()
	}
	return nil
}

func (o *OtelCollector) run() {
	go func() {
		if err := o.collector.Run(o.ctx); err != nil {
			o.logger.Error("collector run failed", zap.Error(err))
		}
	}()
}

func (o *OtelCollector) Provision(ctx caddy.Context) error {
	o.ctx = ctx
	o.logger = ctx.Logger()
	factories, err := components.Components()
	if err != nil {
		return fmt.Errorf("failed to build components: %v", err)
	}
	o.factories = factories
	o.buildInfo = component.BuildInfo{
		Command:     "otelcol-beyond",
		Description: "Custom OpenTelemetry Collector distribution for Beyond project",
		Version:     "0.0.1-dev",
	}
	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{"field:Config"},
			Providers: map[string]confmap.Provider{
				"field": NewProvider(o),
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
	_ beyond.App = (*OtelCollector)(nil)
)
