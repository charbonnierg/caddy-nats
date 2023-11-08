// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package allow

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth/template"
)

// Syntax:
//
//	allow {
//	    account <account_name>
//	    template {
//		     <template>
//	    }
//	}
func (c *AllowAuthCallout) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "template":
				if c.Template == nil {
					c.Template = &template.Template{}
				}
				if err := c.Template.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			case "account":
				if !d.Next() {
					return d.Err("expected account name")
				}
				c.Account = d.Val()
			default:
				return d.Errf("unknown directive '%s'", d.Val())
			}
		}
	}
	return nil
}
