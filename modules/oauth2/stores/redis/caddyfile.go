// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (s *RedisStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if d.NextArg() {
		return d.Err("unexpected positional argument to redis session store")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "connection_url":
			if err := parser.ParseString(d, &s.ConnectionURL); err != nil {
				return err
			}
		case "password":
			if err := parser.ParseString(d, &s.Password); err != nil {
				return err
			}
		case "use_sentinel":
			if err := parser.ParseBool(d, &s.UseSentinel); err != nil {
				return err
			}
		case "sentinel_password":
			if err := parser.ParseString(d, &s.SentinelPassword); err != nil {
				return err
			}
		case "sentinel_master_name":
			if err := parser.ParseString(d, &s.SentinelMasterName); err != nil {
				return err
			}
		case "sentinel_connection_urls":
			if err := parser.ParseStringArray(d, &s.SentinelConnectionURLs); err != nil {
				return err
			}
		case "use_cluster":
			if err := parser.ParseBool(d, &s.UseCluster); err != nil {
				return err
			}
		case "cluster_connection_urls":
			if err := parser.ParseStringArray(d, &s.ClusterConnectionURLs); err != nil {
				return err
			}
		case "ca_path":
			if err := parser.ParseString(d, &s.CAPath); err != nil {
				return err
			}
		case "insecure_skip_tls_verify":
			if err := parser.ParseBool(d, &s.InsecureSkipTLSVerify); err != nil {
				return err
			}
		case "idle_timeout":
			if err := parser.ParseInt(d, &s.IdleTimeout); err != nil {
				return err
			}
		default:
			return d.Err("unrecognized subdirective: " + d.Val())
		}
	}
	return nil
}
