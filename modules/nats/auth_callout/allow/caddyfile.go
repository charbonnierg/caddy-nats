// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package allow

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats/natsapp"
)

func (s *AllowAuthCallout) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "template":
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					val := d.Val()
					switch val {
					case "allow_resp":
						makeTemplate(s)
						s.Template.Resp = &jwt.ResponsePermission{}
					case "allow_pub":
						makeTemplate(s)
						if !d.Next() {
							return d.Err("expected an nats subject")
						}
						s.Template.Pub.Allow = append(s.Template.Pub.Allow, d.Val())
					case "allow_sub":
						makeTemplate(s)
						if !d.Next() {
							return d.Err("expected an nats subject")
						}
						s.Template.Sub.Allow = append(s.Template.Sub.Allow, d.Val())
					case "deny_pub":
						makeTemplate(s)
						if !d.Next() {
							return d.Err("expected an nats subject")
						}
						s.Template.Pub.Deny = append(s.Template.Pub.Deny, d.Val())
					case "deny_sub":
						makeTemplate(s)
						if !d.Next() {
							return d.Err("expected an nats subject")
						}
						s.Template.Sub.Deny = append(s.Template.Sub.Deny, d.Val())
					default:
						return d.Errf("unknown directive '%s'", val)
					}
				}
			case "account":
				if !d.Next() {
					return d.Err("expected account name")
				}
				s.Account = d.Val()
			default:
				return d.Errf("unknown directive '%s'", d.Val())
			}
		}
	}
	return nil
}

func makeTemplate(s *AllowAuthCallout) {
	if s.Template == nil {
		s.Template = &natsapp.Template{}
	}
}
