// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/modules/secrets/automation"
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
		Name:  secrets.NS,
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		automations := []json.RawMessage{}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "store":
				if !d.NextArg() {
					return d.Err("expected a store name")
				}
				storeName := d.Val()
				if !d.NextArg() {
					return d.Err("expected a store type")
				}
				storeTypeShort := d.Val()
				storeType := "secrets.stores." + storeTypeShort
				mod, err := caddyfile.UnmarshalModule(d, storeType)
				if err != nil {
					return d.Errf("failed to unmarshal module '%s': %v", storeType, err)
				}
				if a.StoresRaw == nil {
					a.StoresRaw = make(map[string]json.RawMessage)
				}
				a.StoresRaw[storeName] = caddyconfig.JSONModuleObject(mod, "module", storeTypeShort, nil)
			case "automate", "automation":
				auto := new(automation.Automation)
				err := auto.UnmarshalCaddyfile(d)
				if err != nil {
					return err
				}
				automations = append(automations, caddyconfig.JSON(auto, nil))
			default:
				return d.Errf("unknown secrets subdirective '%s'", d.Val())
			}
		}
		a.AutomationsRaw = append(a.AutomationsRaw, automations...)
	}
	return nil
}
