// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package matchers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func (ConnectionTypeMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.type",
		New: func() caddy.Module { return new(ConnectionTypeMatcher) },
	}
}

type ConnectionTypeMatcher []string

func (m *ConnectionTypeMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *ConnectionTypeMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	typ := strings.ToLower(claims.ClientInformation.Type)
	for _, value := range *m {
		if value == typ {
			return true, nil
		}
	}
	return false, nil
}

func (m *ConnectionTypeMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = ConnectionTypeMatcher{strings.ToLower(v)}
	case []interface{}:
		*m = ConnectionTypeMatcher{}
		for _, value := range v {
			value, ok := value.(string)
			if !ok {
				return fmt.Errorf("connection type must be a string or an array of string")
			}
			if value == "" {
				return fmt.Errorf("connection type cannot be empty")
			}
			*m = append(*m, strings.ToLower(value))
		}
	default:
		return fmt.Errorf("connection type must be a string or an array of string")
	}
	return nil
}

func (u *ConnectionTypeMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = ConnectionTypeMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("connection type cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
