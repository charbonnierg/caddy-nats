// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package changestreamewriter

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

// Syntax:
//
//	destination mongodb_change_stream {
//	    uri <string>
//	    database <string>
//	}
func (e *ChangeStreamWriter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "uri":
			if err := parser.ParseString(d, &e.Uri); err != nil {
				return err
			}
		case "database":
			if err := parser.ParseString(d, &e.Database); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
