// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package otelcol

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	col := new(OtelCollector)
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

func (a *OtelCollector) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	cfg := parseJSON(d)
	a.Config = caddyconfig.JSON(cfg, nil)
	return nil
}

func parseJSON(d *caddyfile.Dispenser) map[string]any {
	cfg := map[string]any{}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		switch d.CountRemainingArgs() {
		case 0:
			cfg[key] = parseJSON(d)
		case 1:
			d.Next()
			if key == "receivers" || key == "processors" || key == "exporters" {
				cfg[key] = []string{d.Val()}
			} else {
				cfg[key] = d.Val()
			}
		default:
			values := []string{}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					values = append(values, val)
				}
			}
			cfg[key] = values
		}
	}
	return cfg
}
