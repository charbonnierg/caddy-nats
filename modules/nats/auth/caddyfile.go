package auth

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

// AuthServiceConfig is the configuration for the auth callout service.
type AuthServiceConfig struct {
	ClientRaw         *natsutils.Client           `json:"client"`
	InternalAccount   string                      `json:"internal_account,omitempty"`
	InternalUser      string                      `json:"internal_user,omitempty"`
	AuthAccount       string                      `json:"auth_account,omitempty"`
	AuthSigningKey    string                      `json:"auth_signing_key,omitempty"`
	SubjectRaw        string                      `json:"subject,omitempty"`
	Policies          policies.ConnectionPolicies `json:"policies,omitempty"`
	DefaultHandlerRaw json.RawMessage             `json:"handler,omitempty" caddy:"namespace=nats.auth_callout inline_key=module"`
}

// UnmarshalCaddyfile sets up the auth config from Caddyfile tokens.
func (s *AuthServiceConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
			err := natscaddyfile.ParseClient(d, s.ClientRaw)
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
