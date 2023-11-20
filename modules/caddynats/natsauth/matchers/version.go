// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func (VersionMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.version",
		New: func() caddy.Module { return new(VersionMatcher) },
	}
}

type VersionMatcher []string

func (m *VersionMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m VersionMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, version := range m {
		if version == claims.ConnectOptions.Version {
			return true, nil
		}
	}
	return false, nil
}

func (m *VersionMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = VersionMatcher{v}
	case []interface{}:
		*m = VersionMatcher{}
		for _, name := range v {
			name, ok := name.(string)
			if !ok {
				return fmt.Errorf("version must be a string or an array of string")
			}
			if name == "" {
				return fmt.Errorf("version cannot be empty")
			}
			*m = append(*m, name)
		}
	default:
		return fmt.Errorf("version must be a string or an array of string")
	}
	return nil
}

func (u *VersionMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = VersionMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("version cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
