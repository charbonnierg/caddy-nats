package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth"
	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

// UnmarshalCaddyfile sets up the auth config from Caddyfile tokens.
func ParseAuthServiceConfig(d *caddyfile.Dispenser, s *auth.AuthServiceConfig) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "internal_account":
			if !d.Args(&s.InternalAccount) {
				return d.Err("expected internal account")
			}
		case "internal_user":
			if !d.Args(&s.InternalUser) {
				return d.Err("expected internal user")
			}
		case "client":
			s.ClientRaw = &natsutils.Client{}
			err := ParseClient(d, s.ClientRaw)
			if err != nil {
				return err
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
			pol := policies.ConnectionPolicy{}
			if err := pol.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			s.Policies = append(s.Policies, &pol)
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
