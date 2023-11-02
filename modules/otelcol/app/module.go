// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	"github.com/quara-dev/beyond/modules/otelcol/app/service"
	"github.com/quara-dev/beyond/modules/otelcol/components"
	"github.com/quara-dev/beyond/modules/otelcol/fieldprovider"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/quara-dev/beyond/modules/otelcol/app/exporters/otlphttp"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/extensions/basicauth"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/extensions/zpages"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/processors/attributes"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/processors/batch"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/receivers/filelog"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/receivers/hostmetrics"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/receivers/otlp"
	_ "github.com/quara-dev/beyond/modules/otelcol/app/receivers/prometheus"
)

func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("otelcol", parseGlobalOption)
	httpcaddyfile.RegisterGlobalOption("telemetry", parseGlobalOption)
}

var (
	BUILD_INFO = component.BuildInfo{
		Command:     "otelcol-beyond",
		Description: "Custom OpenTelemetry Collector distribution for Beyond project",
		Version:     "0.0.1-dev",
	}
)

type cfg struct {
	Raw json.RawMessage
}

type App struct {
	ctx        caddy.Context
	uri        []string
	config     *config.Config
	collector  *otelcol.Collector
	logger     *zap.Logger
	ConfigUri  []string                         `json:"config_uri,omitempty"`
	Service    *service.Service                 `json:"service,omitempty"`
	Receivers  map[component.ID]json.RawMessage `json:"receivers,omitempty"`
	Processors map[component.ID]json.RawMessage `json:"processors,omitempty"`
	Exporters  map[component.ID]json.RawMessage `json:"exporters,omitempty"`
	Extensions map[component.ID]json.RawMessage `json:"extensions,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol",
		New: func() caddy.Module { return new(App) },
	}
}

func (o *App) Provision(ctx caddy.Context) error {
	o.ctx = ctx
	o.logger = ctx.Logger()
	repl := caddy.NewReplacer()
	factories, err := components.Components()
	if err != nil {
		return fmt.Errorf("failed to build components: %v", err)
	}
	o.config = &config.Config{
		Receivers:  map[component.ID]interface{}{},
		Processors: map[component.ID]interface{}{},
		Exporters:  map[component.ID]interface{}{},
		Extensions: map[component.ID]interface{}{},
		Service:    o.Service,
	}
	switch {
	case o.ConfigUri != nil && o.Service == nil:
		o.uri = o.ConfigUri
	case o.ConfigUri == nil && o.Service != nil:
		o.uri = []string{"field:Raw"}
		if err := o.loadRawConfig(o.config); err != nil {
			return err
		}
	default:
		return fmt.Errorf("config_uri and service are mutually exclusive")
	}
	// Register beyond module
	b, err := beyond.Register(o.ctx, o)
	if err != nil {
		return err
	}
	unm, err := b.LoadApp("secrets")
	if err != nil {
		return err
	}
	secretApp, ok := unm.(secrets.App)
	if !ok {
		return fmt.Errorf("expected secrets module")
	}
	secretApp.AddSecretsReplacerVars(repl)
	raw, err := o.config.Marshal(repl)
	if err != nil {
		return err
	}
	provider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: o.uri,
			Providers: map[string]confmap.Provider{
				"field": fieldprovider.NewProvider(&cfg{Raw: raw}),
				"file":  fileprovider.New(),
				"http":  httpsprovider.New(),
				"https": httpsprovider.New(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create config provider: %v", err)
	}
	settings := otelcol.CollectorSettings{
		BuildInfo:               BUILD_INFO,
		Factories:               factories,
		DisableGracefulShutdown: true,
		ConfigProvider:          provider,
		LoggingOptions: []zap.Option{zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return o.logger.Core()
		})},
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

func (a *App) loadRawConfig(dest *config.Config) error {
	if a.Exporters != nil {
		for id, raw := range a.Exporters {
			unm, err := a.ctx.LoadModuleByID("otelcol.exporters."+string(id.Type()), raw)
			if err != nil {
				return err
			}
			_, ok := unm.(config.Exporter)
			if !ok {
				return fmt.Errorf("expected exporter module")
			}
			dest.Exporters[id] = unm
		}
	}
	if a.Processors != nil {
		for id, raw := range a.Processors {
			unm, err := a.ctx.LoadModuleByID("otelcol.processors."+string(id.Type()), raw)
			if err != nil {
				return err
			}
			_, ok := unm.(config.Processor)
			if !ok {
				return fmt.Errorf("expected processor module")
			}
			dest.Processors[id] = unm
		}
	}
	if a.Receivers != nil {
		for id, raw := range a.Receivers {
			unm, err := a.ctx.LoadModuleByID("otelcol.receivers."+string(id.Type()), raw)
			if err != nil {
				return err
			}
			_, ok := unm.(config.Receiver)
			if !ok {
				return fmt.Errorf("expected receiver module")
			}
			dest.Receivers[id] = unm
		}
	}
	if a.Extensions != nil {
		for id, raw := range a.Extensions {
			unm, err := a.ctx.LoadModuleByID("otelcol.extensions."+string(id.Type()), raw)
			if err != nil {
				return err
			}
			_, ok := unm.(config.Extension)
			if !ok {
				return fmt.Errorf("expected extension module")
			}
			dest.Extensions[id] = unm
		}
	}
	return nil
}

func (o *App) Validate() error {
	return nil
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

func (o *App) run() {
	go func() {
		if err := o.collector.Run(o.ctx); err != nil {
			o.logger.Error("collector run failed", zap.Error(err))
		}
	}()
}

var (
	_ beyond.App = (*App)(nil)
)
