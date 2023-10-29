// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure_keyvault

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/azutils"
)

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
				s.CredentialConfig = new(azutils.CredentialConfig)
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
