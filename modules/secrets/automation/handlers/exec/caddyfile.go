// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (h *ExecHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if h.Args == nil {
			h.Args = []string{}
		}
		if d.NextArg() {
			h.Command = d.Val()
			if err := parser.ParseStringArray(d, &h.Args, parser.AllowEmpty()); err != nil {
				return err
			}
		} else {
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "command":
					if !d.NextArg() {
						return d.Err("expected a command")
					}
					h.Command = d.Val()
					if err := parser.ParseStringArray(d, &h.Args, parser.AllowEmpty()); err != nil {
						return err
					}
				case "args":
					if err := parser.ParseStringArray(d, &h.Args); err != nil {
						return err
					}
				default:
					return d.Errf("unknown exec handler property '%s'", d.Val())
				}
			}
		}
	}
	return nil
}
