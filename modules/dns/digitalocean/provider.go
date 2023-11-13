// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package digitalocean

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/digitalocean"
	"github.com/quara-dev/beyond/modules/secrets"
)

// Provider wraps the provider implementation as a Caddy module.
type Provider struct{ *digitalocean.Provider }

func init() {
	caddy.RegisterModule(Provider{})
}

// CaddyModule returns the Caddy module information.
func (Provider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "dns.providers.digitalocean",
		New: func() caddy.Module { return &Provider{new(digitalocean.Provider)} },
	}
}

// Before using the provider config, resolve placeholders in the API token.
// Implements caddy.Provisioner.
func (p *Provider) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()
	if err := secrets.UpdateReplacer(ctx, repl); err != nil {
		return err
	}
	token, err := repl.ReplaceOrErr(p.Provider.APIToken, true, true)
	if err != nil {
		return err
	}
	p.Provider.APIToken = token
	return nil
}

func (p *Provider) Validate() error {
	if p.Provider.APIToken == "" {
		return errors.New("missing digitalocean API token")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*Provider)(nil)
	_ caddy.Provisioner     = (*Provider)(nil)
	_ caddy.Validator       = (*Provider)(nil)
)
