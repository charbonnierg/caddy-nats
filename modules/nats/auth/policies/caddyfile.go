package policies

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats"
)

func (p *ConnectionPolicy) UnmarshalCaddyfileWithAccountName(d *caddyfile.Dispenser, account string) error {
	var connectOpts *ConnectOptsMatcher
	var clientInfo *ClientInfoMatcher
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "match":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "token", "client_token":
					if connectOpts == nil {
						connectOpts = new(ConnectOptsMatcher)
					}
					if !d.NextArg() {
						return d.Err("expected token")
					}
					connectOpts.Token = d.Val()
				case "username", "client_username":
					if connectOpts == nil {
						connectOpts = new(ConnectOptsMatcher)
					}
					if !d.NextArg() {
						return d.Err("expected username")
					}
					connectOpts.User = d.Val()
				case "password", "client_password":
					if connectOpts == nil {
						connectOpts = new(ConnectOptsMatcher)
					}
					if !d.NextArg() {
						return d.Err("expected password")
					}
					connectOpts.Password = d.Val()
				case "client_name", "connection_name":
					if connectOpts == nil {
						connectOpts = new(ConnectOptsMatcher)
					}
					if !d.NextArg() {
						return d.Err("expected name")
					}
					connectOpts.Name = d.Val()
				case "type", "connection_type":
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
				case "kind", "connection_kind":
					if clientInfo == nil {
						clientInfo = new(ClientInfoMatcher)
					}
					if !d.Next() {
						return d.Err("expected kind")
					}
					clientInfo.Kind = d.Val()
				case "host", "client_host":
					if clientInfo == nil {
						clientInfo = new(ClientInfoMatcher)
					}
					if !d.Next() {
						return d.Err("expected host")
					}
					clientInfo.Host = d.Val()
				default:
					return d.Errf("unknown matcher '%s'", d.Val())
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
			callout, ok := mod.(nats.AuthCallout)
			if !ok {
				return d.Errf("module '%s' is not an auth callout", typ)
			}
			if account != "" {
				if err := callout.SetAccount(account); err != nil {
					return err
				}
			}
			p.HandlerRaw = caddyconfig.JSONModuleObject(mod, "module", typ, nil)
		default:
			return d.Errf("unknown policy target '%s'", d.Val())
		}
	}
	if connectOpts != nil {
		if len(p.MatchersRaw) == 0 {
			p.MatchersRaw = []map[string]json.RawMessage{
				{},
			}
		}
		p.MatchersRaw[0]["connect_opts"] = caddyconfig.JSON(connectOpts, nil)
	}
	if clientInfo != nil {
		if len(p.MatchersRaw) == 0 {
			p.MatchersRaw = []map[string]json.RawMessage{
				{},
			}
		}
		p.MatchersRaw[0]["client_info"] = caddyconfig.JSON(clientInfo, nil)
	}
	return nil
}

func (p *ConnectionPolicy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return p.UnmarshalCaddyfileWithAccountName(d, "")
}
