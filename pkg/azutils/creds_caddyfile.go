package azutils

import "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"

func (c *CredentialConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "client_id":
			if !d.AllArgs(&c.ClientId) {
				return d.ArgErr()
			}
		case "client_id_file":
			if !d.AllArgs(&c.ClientIdFile) {
				return d.ArgErr()
			}
		case "client_secret":
			if !d.AllArgs(&c.ClientSecret) {
				return d.ArgErr()
			}
		case "client_secret_file":
			if !d.AllArgs(&c.ClientSecretFile) {
				return d.ArgErr()
			}
		case "tenant_id":
			if !d.AllArgs(&c.TenantId) {
				return d.ArgErr()
			}
		case "tenant_id_file":
			if !d.AllArgs(&c.TenantIdFile) {
				return d.ArgErr()
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
