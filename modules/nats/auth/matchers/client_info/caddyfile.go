package client_info

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (m *ClientInfoMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "host", "ip_address":
			if err := parser.ParseString(d, &m.Host); err != nil {
				return err
			}
		case "kind":
			if err := parser.ParseString(d, &m.Kind); err != nil {
				return err
			}
		case "user":
			if err := parser.ParseString(d, &m.User); err != nil {
				return err
			}
		case "type":
			if err := parser.ParseString(d, &m.Type); err != nil {
				return err
			}
		case "in_process":
			if err := parser.ParseBool(d, &m.InProcess); err != nil {
				return err
			}
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
