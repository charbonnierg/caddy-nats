package natsutils

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
)

func (c *Client) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "internal":
			val, err := caddyutils.ParseBool(d)
			if err != nil {
				return err
			}
			c.Internal = val
		case "name":
			if !d.AllArgs(&c.Name) {
				return d.ArgErr()
			}
		case "servers":
			if c.Servers == nil {
				c.Servers = []string{}
			}
			c.Servers = append(c.Servers, caddyutils.ParseStringArray(d)...)
		case "username":
			if !d.AllArgs(&c.Username) {
				return d.ArgErr()
			}
		case "password":
			if !d.AllArgs(&c.Password) {
				return d.ArgErr()
			}
		case "token":
			if !d.AllArgs(&c.Token) {
				return d.ArgErr()
			}
		case "credentials":
			if !d.AllArgs(&c.Credentials) {
				return d.ArgErr()
			}
		case "seed":
			if !d.AllArgs(&c.Seed) {
				return d.ArgErr()
			}
		case "jwt":
			if !d.AllArgs(&c.Jwt) {
				return d.ArgErr()
			}
		case "jetstream_domain":
			if !d.AllArgs(&c.JSDomain) {
				return d.ArgErr()
			}
		case "jetstream_prefix":
			if !d.AllArgs(&c.JSPrefix) {
				return d.ArgErr()
			}
		case "inbox_prefix":
			if !d.AllArgs(&c.InboxPrefix) {
				return d.ArgErr()
			}
		case "no_randomize":
			val, err := caddyutils.ParseBool(d)
			if err != nil {
				return err
			}
			c.NoRandomize = val
		case "ping_interval":
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			c.PingInterval = val
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
