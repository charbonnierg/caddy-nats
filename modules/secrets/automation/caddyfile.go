// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package automation

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (a *Automation) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	handlers := []json.RawMessage{}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "source":
			a.SourcesRaw = fnutils.DefaultIfEmpty(a.SourcesRaw, []string{})
			if err := caddyutils.ParseStringArray(d, &a.SourcesRaw, false); err != nil {
				return err
			}
		case "template":
			if err := caddyutils.ParseString(d, &a.TemplateRaw); err != nil {
				return err
			}
		case "interval":
			mod, err := caddyfile.UnmarshalModule(d, "secrets.triggers.periodic")
			if err != nil {
				return d.Errf("failed to unmarshal module 'secrets.triggers.periodic': %v", err)
			}
			a.TriggerRaw = caddyconfig.JSONModuleObject(mod, "type", "periodic", nil)
		case "trigger":
			if !d.NextArg() {
				return d.Err("expected a trigger type")
			}
			triggerType := d.Val()
			mod, err := caddyfile.UnmarshalModule(d, "secrets.triggers."+triggerType)
			if err != nil {
				return d.Errf("failed to unmarshal module 'secrets.triggers.%s': %v", triggerType, err)
			}
			a.TriggerRaw = caddyconfig.JSONModuleObject(mod, "type", triggerType, nil)

		case "handle":
			if !d.NextArg() {
				return d.Err("expected a handle type")
			}
			handlerType := d.Val()
			mod, err := caddyfile.UnmarshalModule(d, "secrets.handlers."+handlerType)
			if err != nil {
				return d.Errf("failed to unmarshal module 'secrets.handlers.%s': %v", handlerType, err)
			}
			handlers = append(handlers, caddyconfig.JSONModuleObject(mod, "type", handlerType, nil))
		default:
			return d.Errf("unknown property '%s'", d.Val())
		}
		if len(handlers) > 0 {
			if a.HandlersRaw == nil {
				a.HandlersRaw = []json.RawMessage{}
			}
			a.HandlersRaw = append(a.HandlersRaw, handlers...)
		}
	}
	return nil
}
