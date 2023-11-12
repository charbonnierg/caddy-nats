// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azutils

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (c *CredentialConfig) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "client_id":
			if err := parser.ParseString(d, &c.ClientId); err != nil {
				return err
			}
		case "client_id_file":
			if err := parser.ParseString(d, &c.ClientIdFile); err != nil {
				return err
			}
		case "client_secret":
			if err := parser.ParseString(d, &c.ClientSecret); err != nil {
				return err
			}
		case "client_secret_file":
			if err := parser.ParseString(d, &c.ClientSecretFile); err != nil {
				return err
			}
		case "tenant_id":
			if err := parser.ParseString(d, &c.TenantId); err != nil {
				return err
			}
		case "tenant_id_file":
			if err := parser.ParseString(d, &c.TenantIdFile); err != nil {
				return err
			}
		case "subscription_id":
			if err := parser.ParseString(d, &c.SubscriptionId); err != nil {
				return err
			}
		case "subscription_id_file":
			if err := parser.ParseString(d, &c.SubscriptionIdFile); err != nil {
				return err
			}
		case "access_key":
			if err := parser.ParseString(d, &c.AccessKey); err != nil {
				return err
			}
		case "access_key_file":
			if err := parser.ParseString(d, &c.AccessKeyFile); err != nil {
				return err
			}
		case "no_default_credentials":
			if err := parser.ParseBool(d, &c.NoDefaultCredentials); err != nil {
				return err
			}
		case "no_managed_identity":
			if err := parser.ParseBool(d, &c.NoManagedIdentity); err != nil {
				return err
			}
		case "additionally_allowed_tenant":
			if err := parser.ParseStringArray(d, &c.AdditionallyAllowedTenants); err != nil {
				return err
			}
		case "disable_instance_discovery":
			err := parser.ParseBool(d, &c.DisableInstanceDiscovery)
			if err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
