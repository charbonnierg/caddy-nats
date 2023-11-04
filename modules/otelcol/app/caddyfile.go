// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	"github.com/quara-dev/beyond/modules/otelcol/app/service"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	col := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, col)
		if err != nil {
			return nil, err
		}
	}
	err := col.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "otelcol",
		Value: caddyconfig.JSON(col, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d, parser.Match("otelcol", "telemetry")); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "service":
			if a.Service == nil {
				a.Service = &service.Service{}
			}
			if err := a.Service.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		case "extension":
			name := ""
			if err := parser.ParseString(d, &name); err != nil {
				return err
			}
			id := component.ID{}
			if err := id.UnmarshalText([]byte(name)); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "otelcol.extensions."+string(id.Type()))
			if err != nil {
				return err
			}
			ext, ok := unm.(config.Extension)
			if !ok {
				return d.Errf("expected extension module")
			}
			raw, err := json.Marshal(ext)
			if err != nil {
				return err
			}
			a.Extensions = fnutils.DefaultIfEmptyMap(a.Extensions, map[component.ID]json.RawMessage{})
			a.Extensions[id] = raw
		case "extensions":
			if err := a.UnmarshalExtensions(d); err != nil {
				return err
			}
		case "receiver":
			name := ""
			if err := parser.ParseString(d, &name); err != nil {
				return err
			}
			id := component.ID{}
			if err := id.UnmarshalText([]byte(name)); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "otelcol.receivers."+string(id.Type()))
			if err != nil {
				return err
			}
			rec, ok := unm.(config.Receiver)
			if !ok {
				return d.Errf("expected receiver module")
			}
			raw, err := json.Marshal(rec)
			if err != nil {
				return err
			}
			a.Receivers = fnutils.DefaultIfEmptyMap(a.Receivers, map[component.ID]json.RawMessage{})
			a.Receivers[id] = raw
		case "receivers":
			if err := a.UnmarshalReceivers(d); err != nil {
				return err
			}
		case "processor":
			name := ""
			if err := parser.ParseString(d, &name); err != nil {
				return err
			}
			id := component.ID{}
			if err := id.UnmarshalText([]byte(name)); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "otelcol.processors."+string(id.Type()))
			if err != nil {
				return err
			}
			proc, ok := unm.(config.Processor)
			if !ok {
				return d.Errf("expected processor module")
			}
			raw, err := json.Marshal(proc)
			if err != nil {
				return err
			}
			a.Processors = fnutils.DefaultIfEmptyMap(a.Processors, map[component.ID]json.RawMessage{})
			a.Processors[id] = raw
		case "processors":
			if err := a.UnmarshalProcessors(d); err != nil {
				return err
			}
		case "exporter":
			name := ""
			if err := parser.ParseString(d, &name); err != nil {
				return err
			}
			id := component.ID{}
			if err := id.UnmarshalText([]byte(name)); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "otelcol.exporters."+string(id.Type()))
			if err != nil {
				return err
			}
			exp, ok := unm.(config.Exporter)
			if !ok {
				return d.Errf("expected exporter module")
			}
			raw, err := json.Marshal(exp)
			if err != nil {
				return err
			}
			a.Exporters = fnutils.DefaultIfEmptyMap(a.Exporters, map[component.ID]json.RawMessage{})
			a.Exporters[id] = raw
		case "exporters":
			if err := a.UnmarshalExporters(d); err != nil {
				return err
			}
		default:
			return d.Errf("unknown subdirective %s", d.Val())
		}
	}
	return nil
}

func (a *App) UnmarshalExtensions(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		id := component.ID{}
		if err := id.UnmarshalText([]byte(name)); err != nil {
			return err
		}
		unm, err := caddyfile.UnmarshalModule(d, "otelcol.extensions."+string(id.Type()))
		if err != nil {
			return err
		}
		ext, ok := unm.(config.Extension)
		if !ok {
			return d.Errf("expected extension module")
		}
		raw, err := json.Marshal(ext)
		if err != nil {
			return err
		}
		a.Extensions = fnutils.DefaultIfEmptyMap(a.Extensions, map[component.ID]json.RawMessage{})
		a.Extensions[id] = raw
	}
	return nil
}

func (a *App) UnmarshalReceivers(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		id := component.ID{}
		if err := id.UnmarshalText([]byte(name)); err != nil {
			return err
		}
		unm, err := caddyfile.UnmarshalModule(d, "otelcol.receivers."+string(id.Type()))
		if err != nil {
			return err
		}
		rec, ok := unm.(config.Receiver)
		if !ok {
			return d.Errf("expected receiver module")
		}
		raw, err := json.Marshal(rec)
		if err != nil {
			return d.Errf("failed to marshal receiver: %w", err)
		}
		a.Receivers = fnutils.DefaultIfEmptyMap(a.Receivers, map[component.ID]json.RawMessage{})
		a.Receivers[id] = raw
	}
	return nil
}

func (a *App) UnmarshalProcessors(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		id := component.ID{}
		if err := id.UnmarshalText([]byte(name)); err != nil {
			return err
		}
		unm, err := caddyfile.UnmarshalModule(d, "otelcol.processors."+string(id.Type()))
		if err != nil {
			return err
		}
		proc, ok := unm.(config.Processor)
		if !ok {
			return d.Errf("expected processor module")
		}
		raw, err := json.Marshal(proc)
		if err != nil {
			return err
		}
		a.Processors = fnutils.DefaultIfEmptyMap(a.Processors, map[component.ID]json.RawMessage{})
		a.Processors[id] = raw
	}
	return nil
}

func (a *App) UnmarshalExporters(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		id := component.ID{}
		if err := id.UnmarshalText([]byte(name)); err != nil {
			return err
		}
		unm, err := caddyfile.UnmarshalModule(d, "otelcol.exporters."+string(id.Type()))
		if err != nil {
			return err
		}
		exp, ok := unm.(config.Exporter)
		if !ok {
			return d.Errf("expected exporter module")
		}
		raw, err := json.Marshal(exp)
		if err != nil {
			return err
		}
		a.Exporters = fnutils.DefaultIfEmptyMap(a.Exporters, map[component.ID]json.RawMessage{})
		a.Exporters[id] = raw
	}
	return nil
}
