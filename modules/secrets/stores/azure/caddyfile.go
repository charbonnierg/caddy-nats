// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure

import "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"

func (s *AzureKeyvault) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			s.URI = d.Val()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "uri":
				if !d.AllArgs(&s.URI) {
					return d.Err("expected a single argument for uri directive")
				}
			case "creds":
				s.CredentialConfig = &AzCredentialConfig{}
				if err := s.CredentialConfig.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

func (c *AzCredentialConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "client_id":
			if !d.AllArgs(&c.ClientId) {
				return d.ArgErr()
			}
		case "client_id_file":
			if !d.AllArgs(&c.ClientIdFile) {
				return d.ArgErr()
			}
		case "client_secret":
			if !d.AllArgs(&c.ClientSecret) {
				return d.ArgErr()
			}
		case "client_secret_file":
			if !d.AllArgs(&c.ClientSecretFile) {
				return d.ArgErr()
			}
		case "tenant_id":
			if !d.AllArgs(&c.TenantId) {
				return d.ArgErr()
			}
		case "tenant_id_file":
			if !d.AllArgs(&c.TenantIdFile) {
				return d.ArgErr()
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
