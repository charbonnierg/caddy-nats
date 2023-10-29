package azutils

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
)

func (c *CredentialConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "client_id":
			if err := caddyutils.ParseString(d, &c.ClientId); err != nil {
				return err
			}
		case "client_id_file":
			if err := caddyutils.ParseString(d, &c.ClientIdFile); err != nil {
				return err
			}
		case "client_secret":
			if err := caddyutils.ParseString(d, &c.ClientSecret); err != nil {
				return err
			}
		case "client_secret_file":
			if err := caddyutils.ParseString(d, &c.ClientSecretFile); err != nil {
				return err
			}
		case "tenant_id":
			if err := caddyutils.ParseString(d, &c.TenantId); err != nil {
				return err
			}
		case "tenant_id_file":
			if err := caddyutils.ParseString(d, &c.TenantIdFile); err != nil {
				return err
			}
		case "no_default_credentials":
			if err := caddyutils.ParseBool(d, &c.NoDefaultCredentials); err != nil {
				return err
			}
		case "no_managed_identity":
			if err := caddyutils.ParseBool(d, &c.NoManagedIdentity); err != nil {
				return err
			}
		case "additionally_allowed_tenant":
			if err := caddyutils.ParseStringArray(d, &c.AdditionallyAllowedTenants, false); err != nil {
				return err
			}
		case "disable_instance_discovery":
			err := caddyutils.ParseBool(d, &c.DisableInstanceDiscovery)
			if err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
