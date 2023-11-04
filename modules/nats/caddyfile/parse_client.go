package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

func ParseClient(d *caddyfile.Dispenser, c *natsutils.Client) error {
	if d.NextArg() {
		switch d.Val() {
		case "in_process":
			c.Internal = true
		default:
			c.Name = d.Val()
		}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "in_process", "internal":
			if err := parser.ParseBool(d, &c.Internal); err != nil {
				return err
			}
		case "name":
			if err := parser.ParseString(d, &c.Name); err != nil {
				return err
			}
		case "servers":
			if err := parser.ParseStringArray(d, &c.Servers); err != nil {
				return err
			}
		case "username":
			if err := parser.ParseString(d, &c.Username); err != nil {
				return err
			}
		case "password":
			if err := parser.ParseString(d, &c.Password); err != nil {
				return err
			}
		case "token":
			if err := parser.ParseString(d, &c.Token); err != nil {
				return err
			}
		case "credentials":
			if err := parser.ParseString(d, &c.Credentials); err != nil {
				return err
			}
		case "seed":
			if err := parser.ParseString(d, &c.Seed); err != nil {
				return err
			}
		case "jwt":
			if err := parser.ParseString(d, &c.Jwt); err != nil {
				return err
			}
		case "jetstream_domain":
			if err := parser.ParseString(d, &c.JSDomain); err != nil {
				return err
			}
		case "jetstream_prefix":
			if err := parser.ParseString(d, &c.JSPrefix); err != nil {
				return err
			}
		case "inbox_prefix":
			if err := parser.ParseString(d, &c.InboxPrefix); err != nil {
				return err
			}
		case "no_randomize":
			if err := parser.ParseBool(d, &c.NoRandomize); err != nil {
				return err
			}
		case "ping_interval":
			if err := parser.ParseDuration(d, &c.PingInterval); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
