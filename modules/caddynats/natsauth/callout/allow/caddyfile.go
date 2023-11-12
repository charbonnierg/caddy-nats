// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package allow

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

// Syntax:
//
//	allow {
//	  <template>
//	}
func (c *AllowAuthCallout) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // skip "allow"
	c.Template = &natsauth.Template{}
	if err := c.Template.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	return nil
}
