// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package http_handler

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/quara-dev/beyond/modules/oauth2/oauth2app"
)

// ParsePublishHandler parses the nats_publish directive. Syntax:
//
//	oauth2_session {
//
// }
func ParseOauth2ProxyDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var p = OAuth2Session{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

func (p *OAuth2Session) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if p.EndpointRaw == nil {
				p.EndpointRaw = &oauth2app.Endpoint{}
			}
			if !d.Args(&p.EndpointRaw.Name) {
				return d.Err("expected endpoint name")
			}
			if err := p.EndpointRaw.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		}
	}
	return nil
}
