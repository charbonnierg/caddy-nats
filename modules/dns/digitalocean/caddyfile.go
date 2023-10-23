// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package digitalocean

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// UnmarshalCaddyfile sets up the DNS provider from Caddyfile tokens. Syntax:
//
//	digitalocean [<api_token>] {
//	    api_token <api_token>
//	}
func (p *Provider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			p.Provider.APIToken = d.Val()
		}
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "api_token":
				if p.Provider.APIToken != "" {
					return d.Err("API token already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.Provider.APIToken = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}
			case "api_token_file":
				if p.Provider.APIToken != "" {
					return d.Err("API token already set")
				}
				if !d.NextArg() {
					return d.ArgErr()
				}
				path := d.Val()
				if path == "~" {
					// In case of "~", which won't be caught by the "else if"
					usr, err := user.Current()
					if err != nil {
						return d.Errf("resolving user home directory: %v", err)
					}
					path = usr.HomeDir
				} else if strings.HasPrefix(path, "~/") {
					usr, err := user.Current()
					if err != nil {
						return d.Errf("resolving user home directory: %v", err)
					}
					path = filepath.Join(usr.HomeDir, path[2:])
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return d.Errf("reading API token file: %v", err)
				}
				p.Provider.APIToken = string(data)
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	if p.Provider.APIToken == "" {
		return d.Err("missing API token")
	}
	return nil
}
