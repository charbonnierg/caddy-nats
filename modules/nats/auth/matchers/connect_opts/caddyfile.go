package connect_opts

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (m *ConnectOptsMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "token", "client_token":
			if err := parser.ParseString(d, &m.Token); err != nil {
				return err
			}
		case "user", "username", "client_username":
			if err := parser.ParseString(d, &m.User); err != nil {
				return err
			}
		case "pass", "password", "client_password":
			if err := parser.ParseString(d, &m.Password); err != nil {
				return err
			}
		case "client_name", "connection_name":
			if err := parser.ParseString(d, &m.Name); err != nil {
				return err
			}
		case "lang":
			if err := parser.ParseString(d, &m.Lang); err != nil {
				return err
			}
		case "version":
			if err := parser.ParseString(d, &m.Version); err != nil {
				return err
			}
		case "protocol":
			if err := parser.ParseInt(d, &m.Protocol); err != nil {
				return err
			}
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
