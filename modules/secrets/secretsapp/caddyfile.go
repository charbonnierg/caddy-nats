// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	a := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, a)
		if err != nil {
			return nil, err
		}
	}
	err := a.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "secrets",
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "stores":
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					storeName := d.Val()
					if !d.NextArg() {
						return d.Err("expected a store type")
					}
					storeType := "secrets.stores." + d.Val()
					mod, err := caddyfile.UnmarshalModule(d, storeType)
					if err != nil {
						return d.Errf("failed to unmarshal module '%s': %v", storeType, err)
					}
					if a.StoresRaw == nil {
						a.StoresRaw = make(map[string]json.RawMessage)
					}
					a.StoresRaw[storeName] = caddyconfig.JSONModuleObject(mod, "type", storeType, nil)
				}
			}
		}
	}
	return nil
}
