package changestreamexporter

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
func (e *ChangeStreamExporter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
