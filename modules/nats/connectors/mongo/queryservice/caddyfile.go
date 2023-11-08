package mongoservice

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (s *MongoQueryService) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "uri":
			if err := parser.ParseString(d, &s.Uri); err != nil {
				return err
			}
		case "queue_group":
			if err := parser.ParseString(d, &s.QueueGroup); err != nil {
				return err
			}
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
