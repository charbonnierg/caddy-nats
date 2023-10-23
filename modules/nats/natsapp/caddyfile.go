// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/modules/nats/internal/natsoptions"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	a := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, a)
		if err != nil {
			return nil, err
		}
	}
	err := a.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "nats",
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "auth_service":
				if a.AuthService == nil {
					a.AuthService = new(AuthService)
				}
				if err := a.AuthService.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			case "server":
				if a.Options == nil {
					a.Options = new(natsoptions.Options)
				}
				if err := a.Options.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			default:
				return d.Errf("unknown directive '%s'", d.Val())
			}
		}
	}
	return nil
}

func (s *AuthService) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "internal_account":
			if !d.Args(&s.InternalAccount) {
				return d.Err("expected internal account")
			}
		case "default":
			if !d.NextArg() {
				return d.Err("expected default handler")
			}
			typ := d.Val()
			mod, err := caddyfile.UnmarshalModule(d, "nats.auth_callout."+typ)
			if err != nil {
				return d.Errf("failed to unmarshal module '%s': %v", typ, err)
			}
			s.DefaultHandlerRaw = caddyconfig.JSONModuleObject(mod, "module", typ, nil)
		case "policy":
			pol := ConnectionPolicy{}
			var connectOpts *ConnectOptsMatcher
			var clientInfo *ClientInfoMatcher
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "match":
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "username":
							if connectOpts == nil {
								connectOpts = new(ConnectOptsMatcher)
							}
							if !d.NextArg() {
								return d.Err("expected username")
							}
							connectOpts.User = d.Val()
						case "password":
							if connectOpts == nil {
								connectOpts = new(ConnectOptsMatcher)
							}
							if !d.NextArg() {
								return d.Err("expected password")
							}
							connectOpts.Password = d.Val()
						case "connection_name":
							if connectOpts == nil {
								connectOpts = new(ConnectOptsMatcher)
							}
							if !d.NextArg() {
								return d.Err("expected name")
							}
							connectOpts.Name = d.Val()
						case "connection_type":
							if clientInfo == nil {
								clientInfo = new(ClientInfoMatcher)
							}
							if !d.Next() {
								return d.Err("expected connection type")
							}
							switch d.Val() {
							case "in_process":
								clientInfo.InProcess = true
							default:
								clientInfo.Type = d.Val()
							}
						case "kind":
							if clientInfo == nil {
								clientInfo = new(ClientInfoMatcher)
							}
							if !d.Next() {
								return d.Err("expected kind")
							}
							clientInfo.Kind = d.Val()
						case "host":
							if clientInfo == nil {
								clientInfo = new(ClientInfoMatcher)
							}
							if !d.Next() {
								return d.Err("expected host")
							}
							clientInfo.Host = d.Val()
						}
					}
				case "callout":
					if !d.NextArg() {
						return d.Err("expected callout type")
					}
					typ := d.Val()
					mod, err := caddyfile.UnmarshalModule(d, "nats.auth_callout."+typ)
					if err != nil {
						return d.Errf("failed to unmarshal module '%s': %v", typ, err)
					}
					pol.HandlerRaw = caddyconfig.JSONModuleObject(mod, "module", typ, nil)
				default:
					return d.Errf("unknown policy target '%s'", d.Val())
				}
			}
			if connectOpts != nil {
				if len(pol.MatchersRaw) == 0 {
					pol.MatchersRaw = []map[string]json.RawMessage{
						{},
					}
				}
				pol.MatchersRaw[0]["connect_opts"] = caddyconfig.JSON(connectOpts, nil)
			}
			if clientInfo != nil {
				if len(pol.MatchersRaw) == 0 {
					pol.MatchersRaw = []map[string]json.RawMessage{
						{},
					}
				}
				pol.MatchersRaw[0]["client_info"] = caddyconfig.JSON(clientInfo, nil)
			}
			s.Policies = append(s.Policies, &pol)
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
