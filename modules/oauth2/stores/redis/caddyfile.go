// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func (s *RedisStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if d.NextArg() {
		return d.Err("unexpected positional argument to redis session store")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "connection_url":
			if !d.AllArgs(&s.ConnectionURL) {
				return d.ArgErr()
			}
		case "password":
			if !d.AllArgs(&s.Password) {
				return d.ArgErr()
			}
		case "use_sentinel":
			val, err := parseBool(d)
			if err != nil {
				return err
			}
			s.UseSentinel = val
		case "sentinel_password":
			if !d.AllArgs(&s.SentinelPassword) {
				return d.ArgErr()
			}
		case "sentinel_master_name":
			if !d.AllArgs(&s.SentinelMasterName) {
				return d.ArgErr()
			}
		case "sentinel_connection_urls":
			if s.SentinelConnectionURLs == nil {
				s.SentinelConnectionURLs = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					s.SentinelConnectionURLs = append(s.SentinelConnectionURLs, val)
				}
			}
		case "use_cluster":
			val, err := parseBool(d)
			if err != nil {
				return err
			}
			s.UseCluster = val
		case "cluster_connection_urls":
			if s.ClusterConnectionURLs == nil {
				s.ClusterConnectionURLs = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					s.ClusterConnectionURLs = append(s.ClusterConnectionURLs, val)
				}
			}
		case "ca_path":
			if !d.AllArgs(&s.CAPath) {
				return d.ArgErr()
			}
		case "insecure_skip_tls_verify":
			val, err := parseBool(d)
			if err != nil {
				return err
			}
			s.InsecureSkipTLSVerify = val
		case "idle_timeout":
			val, err := parseInt(d)
			if err != nil {
				return err
			}
			s.IdleTimeout = val
		default:
			return d.Err("unrecognized subdirective: " + d.Val())
		}
	}
	return nil
}

func parseInt(d *caddyfile.Dispenser) (int, error) {
	raw := ""
	if !d.AllArgs(&raw) {
		return 0, d.ArgErr()
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, d.Errf("invalid integer value: %s", raw)
	}
	return val, nil
}

func parseBool(d *caddyfile.Dispenser) (bool, error) {
	if !d.NextArg() {
		return true, nil
	}
	raw := d.Val()
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return false, d.Errf("invalid boolean value: %s", raw)
	}
	return val, nil
}
