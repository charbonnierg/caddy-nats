package auth

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

// UnmarshalCaddyfile sets up the auth config from Caddyfile tokens.
func (a *AuthService) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "internal_account", "account":
			a.Connection = fnutils.DefaultIfNil(a.Connection, &client.Connection{})
			if err := parser.ParseString(d, &a.Connection.Account); err != nil {
				return err
			}
		case "auth_public_key":
			if err := parser.ParseString(d, &a.AuthPublicKey); err != nil {
				return err
			}
		case "signing_key", "auth_signing_key":
			if err := parser.ParseString(d, &a.AuthSigningKey); err != nil {
				return err
			}
		case "client", "connection":
			return d.Err("client block is not supported in auth service")
		case "default":
			if !d.NextArg() {
				return d.Err("expected default handler")
			}
			typ := d.Val()
			mod, err := caddyfile.UnmarshalModule(d, "nats.auth_callout."+typ)
			if err != nil {
				return d.Errf("failed to unmarshal module '%s': %v", typ, err)
			}
			a.DefaultHandlerRaw = caddyconfig.JSONModuleObject(mod, "module", typ, nil)
		case "policy":
			pol := policies.ConnectionPolicy{}
			if err := pol.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			a.Policies = append(a.Policies, &pol)
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	return nil
}
