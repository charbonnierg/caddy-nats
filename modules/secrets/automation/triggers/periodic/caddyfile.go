// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package periodic

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (t *PeriodicTrigger) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if d.NextArg() {
		if err := parser.ParseDuration(d, &t.Interval); err != nil {
			return err
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "interval":
				if err := parser.ParseDuration(d, &t.Interval); err != nil {
					return err
				}
			default:
				return d.Errf("unknown periodic trigger property '%s'", d.Val())
			}
		}
	}
	return nil
}
