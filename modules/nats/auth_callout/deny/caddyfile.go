// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package deny

import "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"

func (s *DenyAuthCallout) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "message":
				if !d.Next() {
					return d.Err("expected message")
				}
				s.Message = d.Val()
			default:
				return d.Errf("unknown directive '%s'", d.Val())
			}
		}
	}
	return nil
}
