// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package eventgridwriter

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/azutils"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (r *AzureEventGridWriter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "endpoint":
			if err := parser.ParseString(d, &r.Endpoint); err != nil {
				return err
			}
		case "event_type":
			if err := parser.ParseString(d, &r.EventType); err != nil {
				return err
			}
		case "topic_name":
			if err := parser.ParseString(d, &r.TopicName); err != nil {
				return err
			}
		case "credential", "credentials":
			creds := &azutils.CredentialConfig{}
			if err := creds.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			r.Credentials = creds
		default:
			return d.Errf("unknown property '%s'", d.Val())
		}
	}
	return nil
}
