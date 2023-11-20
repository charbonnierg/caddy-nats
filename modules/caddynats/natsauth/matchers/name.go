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

func (NameMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.name",
		New: func() caddy.Module { return new(NameMatcher) },
	}
}

type NameMatcher []string

func (m *NameMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m NameMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, username := range m {
		if username == claims.ConnectOptions.Name {
			return true, nil
		}
		if username == claims.ClientInformation.Name {
			return true, nil
		}
	}
	return false, nil
}

func (m *NameMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = NameMatcher{v}
	case []interface{}:
		*m = NameMatcher{}
		for _, name := range v {
			name, ok := name.(string)
			if !ok {
				return fmt.Errorf("username must be a string or an array of string")
			}
			if name == "" {
				return fmt.Errorf("username cannot be empty")
			}
			*m = append(*m, name)
		}
	default:
		return fmt.Errorf("username must be a string or an array of string")
	}
	return nil
}

func (u *NameMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = NameMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("username cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
