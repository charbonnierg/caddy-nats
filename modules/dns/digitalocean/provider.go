// Copyright 2023 Guillaume Charbonnier
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package digitalocean

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/libdns/digitalocean"
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
	p.Provider.APIToken = repl.ReplaceAll(p.Provider.APIToken, "")
	return nil
}

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
