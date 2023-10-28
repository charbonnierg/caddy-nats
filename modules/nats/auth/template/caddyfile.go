package template

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
)

func (t *Template) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		val := d.Val()
		switch val {
		case "allow_resp":
			t.Resp = &jwt.ResponsePermission{}
		case "allow_pub":
			if !d.Next() {
				return d.Err("expected an nats subject")
			}
			t.Pub.Allow = append(t.Pub.Allow, d.Val())
		case "allow_sub":
			if !d.Next() {
				return d.Err("expected an nats subject")
			}
			t.Sub.Allow = append(t.Sub.Allow, d.Val())
		case "deny_pub":
			if !d.Next() {
				return d.Err("expected an nats subject")
			}
			t.Pub.Deny = append(t.Pub.Deny, d.Val())
		case "deny_sub":
			if !d.Next() {
				return d.Err("expected an nats subject")
			}
			t.Sub.Deny = append(t.Sub.Deny, d.Val())
		default:
			return d.Errf("unknown directive '%s'", val)
		}
	}
	return nil
}
