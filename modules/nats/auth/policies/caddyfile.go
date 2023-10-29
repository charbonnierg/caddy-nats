package policies

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func (p *ConnectionPolicy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var connectOpts *ConnectOptsMatcher
	var clientInfo *ClientInfoMatcher
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "match":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "token":
					if connectOpts == nil {
						connectOpts = new(ConnectOptsMatcher)
					}
					if !d.NextArg() {
						return d.Err("expected token")
					}
					connectOpts.Token = d.Val()
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
