// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package changestreamreader

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

// Syntax:
//
//	source mongodb_change_stream {
//	    uri <string>
//	    database <string>
//	    collection <string>
//	    resume_token_database <string>
//	    resume_token_collection <string>
//	}
func (r *ChangeStreamReader) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "uri":
			if err := parser.ParseString(d, &r.Uri); err != nil {
				return err
			}
		case "database":
			if err := parser.ParseString(d, &r.Database); err != nil {
				return err
			}
		case "collection":
			if err := parser.ParseString(d, &r.Collection); err != nil {
				return err
			}
		case "resume_token_database":
			if err := parser.ParseString(d, &r.ResumeTokenDatabase); err != nil {
				return err
			}
		case "resume_token_collection":
			if err := parser.ParseString(d, &r.ResumeTokenCollection); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
