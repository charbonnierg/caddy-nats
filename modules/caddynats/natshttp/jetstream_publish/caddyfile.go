// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream_publish

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func parseHandlerDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	p := &JetStreamPublish{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

// UnmarshalCaddyfile parses the "jetstream_publish" directive from
// a Caddyfile dispenser.
// Syntax:
//
//	  jetstream_publish {
//		   connection <name> <type> {
//		      [options]
//		   }
//		   subject <subject>
//	  }
//
// Example:
//
//	  jetstream_publish {
//		   connection my-client tcp {
//		      urls tls://localhost:4222
//		   }
//		   subject some.subject
//	  }
func (p *JetStreamPublish) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	if d.Val() != "jetstream_publish" {
		return d.Errf("Expected 'jetstream_publish' directive, got '%s'", d.Val())
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "subject":
			sub := ""
			if !d.AllArgs(&sub) {
				return d.Err("invalid subject")
			}
			p.Subject = sub
		case "account":
			account := ""
			if !d.AllArgs(&account) {
				return d.Err("invalid account")
			}
			p.Account = account
		case "client":
			p.Client = fnutils.DefaultIfNil(p.Client, &natsclient.NatsClient{})
			if err := p.Client.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
