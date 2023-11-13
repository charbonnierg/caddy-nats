// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/otelcol/app/service"
	"go.opentelemetry.io/collector/component"
)

type Receiver interface {
	UnmarshalCaddyfile(d *caddyfile.Dispenser) error
	ReplaceAll(repl *caddy.Replacer) error
}

type Processor interface {
	UnmarshalCaddyfile(d *caddyfile.Dispenser) error
	ReplaceAll(repl *caddy.Replacer) error
}

type Exporter interface {
	UnmarshalCaddyfile(d *caddyfile.Dispenser) error
	ReplaceAll(repl *caddy.Replacer) error
}

type Extension interface {
	UnmarshalCaddyfile(d *caddyfile.Dispenser) error
	ReplaceAll(repl *caddy.Replacer) error
}

type Config struct {
	Receivers  map[component.ID]interface{} `json:"receivers,omitempty"`
	Processors map[component.ID]interface{} `json:"processors,omitempty"`
	Exporters  map[component.ID]interface{} `json:"exporters,omitempty"`
	Extensions map[component.ID]interface{} `json:"extensions,omitempty"`
	Service    *service.Service             `json:"service,omitempty"`
}

func (c *Config) Marshal(repl *caddy.Replacer) ([]byte, error) {
	values := map[string]map[string]any{}
	if c.Receivers != nil {
		values["receivers"] = map[string]any{}
		for name, receiver := range c.Receivers {
			rec, ok := receiver.(Receiver)
			if !ok {
				return nil, errors.New("expected exporter module")
			}
			if err := rec.ReplaceAll(repl); err != nil {
				return nil, err
			}
			raw, err := json.Marshal(rec)
			if err != nil {
				return nil, err
			}
			target := map[string]any{}
			if err := json.Unmarshal(raw, &target); err != nil {
				return nil, err
			}
			values["receivers"][name.String()] = target
		}
	}
	if c.Processors != nil {
		values["processors"] = map[string]any{}
		for name, processor := range c.Processors {
			pro, ok := processor.(Processor)
			if !ok {
				return nil, errors.New("expected exporter module")
			}
			if err := pro.ReplaceAll(repl); err != nil {
				return nil, err
			}
			raw, err := json.Marshal(pro)
			if err != nil {
				return nil, err
			}
			target := map[string]any{}
			if err := json.Unmarshal(raw, &target); err != nil {
				return nil, err
			}
			values["processors"][name.String()] = target
		}
	}
	if c.Exporters != nil {
		values["exporters"] = map[string]any{}
		for name, exporter := range c.Exporters {
			exp, ok := exporter.(Exporter)
			if !ok {
				return nil, errors.New("expected exporter module")
			}
			if err := exp.ReplaceAll(repl); err != nil {
				return nil, err
			}
			raw, err := json.Marshal(exp)
			if err != nil {
				return nil, err
			}
			target := map[string]any{}
			if err := json.Unmarshal(raw, &target); err != nil {
				return nil, err
			}
			values["exporters"][name.String()] = target
		}
	}
	if c.Extensions != nil {
		values["extensions"] = map[string]any{}
		for name, extension := range c.Extensions {
			ext, ok := extension.(Extension)
			if !ok {
				return nil, errors.New("expected extension module")
			}
			if err := ext.ReplaceAll(repl); err != nil {
				return nil, err
			}
			raw, err := json.Marshal(ext)
			if err != nil {
				return nil, err
			}
			target := map[string]any{}
			if err := json.Unmarshal(raw, &target); err != nil {
				return nil, err
			}
			values["extensions"][name.String()] = target
		}
	}
	srv, err := c.Service.Values()
	if err != nil {
		return nil, err
	}
	values["service"] = srv
	return json.Marshal(values)
}
