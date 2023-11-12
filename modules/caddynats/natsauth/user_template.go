// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsauth

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
)

type Template jwt.User

func (t *Template) Render(request AuthorizationRequest, user *jwt.UserClaims) {
	repl := request.Replacer()
	// Copy the template
	if t.Payload != 0 {
		user.NatsLimits.Payload = t.Payload
	}
	if t.Subs != 0 {
		user.NatsLimits.Subs = t.Subs
	}
	if t.Data != 0 {
		user.NatsLimits.Data = t.Payload
	}
	if t.Permissions.Pub.Allow != nil {
		user.Permissions.Pub.Allow = jwt.StringList{}
		for _, allow := range t.Permissions.Pub.Allow {
			user.Permissions.Pub.Allow.Add(repl.ReplaceKnown(allow, ""))
		}
	}
	if t.Permissions.Pub.Deny != nil {
		user.Permissions.Pub.Deny = jwt.StringList{}
		for _, deny := range t.Permissions.Pub.Deny {
			user.Permissions.Pub.Deny.Add(repl.ReplaceKnown(deny, ""))
		}
	}
	if t.Permissions.Sub.Allow != nil {
		user.Permissions.Sub.Allow = jwt.StringList{}
		for _, allow := range t.Permissions.Sub.Allow {
			user.Permissions.Sub.Allow.Add(repl.ReplaceKnown(allow, ""))
		}
	}
	if t.Permissions.Sub.Deny != nil {
		user.Permissions.Sub.Deny = jwt.StringList{}
		for _, deny := range t.Permissions.Sub.Deny {
			user.Permissions.Sub.Deny.Add(repl.ReplaceKnown(deny, ""))
		}
	}
	if t.Permissions.Resp != nil {
		user.Permissions.Resp = t.Permissions.Resp
	}
	if t.Times != nil {
		user.UserLimits.Times = t.Times
	}
	if t.Src != nil {
		user.UserLimits.Src = jwt.CIDRList{}
		for _, src := range t.Src {
			user.UserLimits.Src.Add(repl.ReplaceKnown(src, ""))
		}
	}
	if t.Locale != "" {
		user.UserLimits.Locale = repl.ReplaceKnown(t.Locale, "")
	}
	if t.AllowedConnectionTypes != nil {
		user.AllowedConnectionTypes = jwt.StringList{}
		for _, connType := range t.AllowedConnectionTypes {
			user.AllowedConnectionTypes.Add(repl.ReplaceKnown(connType, ""))
		}
	}
	if t.BearerToken {
		user.BearerToken = true
	}
}

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
