package basicauth

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func (r *BasicAuthExtension) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var name string
	if err := parser.ParseString(d, &name); err != nil {
		return err
	}
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return err
	}
	if id.Type() != "basicauth" {
		return d.Errf("expected basicauth extension, got %s", id.Type())
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "username":
			r.ClientAuth = fnutils.DefaultIfNil(r.ClientAuth, &ClientAuth{})
			if err := parser.ParseString(d, &r.ClientAuth.Username); err != nil {
				return err
			}
		case "password":
			r.ClientAuth = fnutils.DefaultIfNil(r.ClientAuth, &ClientAuth{})
			if err := parser.ParseString(d, &r.ClientAuth.Password); err != nil {
				return err
			}
		case "htpasswd_file":
			r.Htpasswd = fnutils.DefaultIfNil(r.Htpasswd, &ServerAuth{})
			if err := parser.ParseString(d, &r.Htpasswd.File); err != nil {
				return err
			}
		case "users":
			r.Htpasswd = fnutils.DefaultIfNil(r.Htpasswd, &ServerAuth{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				user := d.Val()
				var password string
				if err := parser.ParseString(d, &password); err != nil {
					return err
				}
				r.Htpasswd.Users = fnutils.DefaultIfEmptyMap(r.Htpasswd.Users, make(map[string]string))
				r.Htpasswd.Users[user] = password
			}
		case "user":
			r.Htpasswd = fnutils.DefaultIfNil(r.Htpasswd, &ServerAuth{})
			var user string
			var password string
			if err := parser.ParseString(d, &user); err != nil {
				return err
			}
			if err := parser.ParseString(d, &password); err != nil {
				return err
			}
			r.Htpasswd.Users = fnutils.DefaultIfEmptyMap(r.Htpasswd.Users, make(map[string]string))
			r.Htpasswd.Users[user] = password
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
