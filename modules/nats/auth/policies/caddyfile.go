package policies

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (p *ConnectionPolicy) UnmarshalCaddyfileWithAccountName(d *caddyfile.Dispenser, account string) error {

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "match":
			var matcher string
			if err := parser.ParseString(d, &matcher); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "nats.matchers."+matcher)
			if err != nil {
				return err
			}
			m, ok := unm.(nats.Matcher)
			if !ok {
				return d.Errf("module '%s' is not a matcher", matcher)
			}
			p.MatchersRaw = fnutils.DefaultIfEmpty(p.MatchersRaw, []json.RawMessage{})
			p.MatchersRaw = append(p.MatchersRaw, caddyconfig.JSONModuleObject(m, "type", matcher, nil))
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
	return nil
}

func (p *ConnectionPolicy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return p.UnmarshalCaddyfileWithAccountName(d, "")
}
