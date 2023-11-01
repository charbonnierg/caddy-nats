// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package http_handler

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/quara-dev/beyond/modules/oauth2/endpoint"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

// ParsePublishHandler parses the nats_publish directive. Syntax:
//
//	authorize_with {
//
// }
func ParseOauth2ProxyDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var p = OAuth2Session{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

func (p *OAuth2Session) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "authorize_with"); err != nil {
		return err
	}
	if d.CountRemainingArgs() > 0 {
		if err := caddyutils.ExpectString(d, "oauth2"); err != nil {
			return err
		}
		p.EndpointRaw = fnutils.DefaultIfNil(p.EndpointRaw, &endpoint.Endpoint{})
		if err := caddyutils.ParseString(d, &p.EndpointRaw.NameRaw); err != nil {
			return err
		}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		p.EndpointRaw = fnutils.DefaultIfNil(p.EndpointRaw, &endpoint.Endpoint{})
		if err := caddyutils.ParseString(d, &p.EndpointRaw.NameRaw); err != nil {
			return err
		}
		if err := p.EndpointRaw.UnmarshalCaddyfile(d); err != nil {
			return err
		}
	}
	return nil
}
