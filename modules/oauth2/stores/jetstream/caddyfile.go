// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (s *JetStreamStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "name":
				if err := parser.ParseString(d, &s.Name); err != nil {
					return err
				}
			case "ttl":
				if err := parser.ParseDuration(d, &s.TTL); err != nil {
					return err
				}
			case "client":
				s.Client = &natsclient.NatsClient{}
				if err := s.Client.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
