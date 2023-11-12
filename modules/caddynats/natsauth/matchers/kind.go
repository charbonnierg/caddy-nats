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

func (ConnectionKindMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.kind",
		New: func() caddy.Module { return new(ConnectionKindMatcher) },
	}
}

type ConnectionKindMatcher []string

func (m *ConnectionKindMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *ConnectionKindMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	kind := strings.ToLower(claims.ClientInformation.Kind)
	for _, value := range *m {
		if value == kind {
			return true, nil
		}
	}
	return false, nil
}

func (m *ConnectionKindMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = ConnectionKindMatcher{strings.ToLower(v)}
	case []interface{}:
		*m = ConnectionKindMatcher{}
		for _, value := range v {
			value, ok := value.(string)
			if !ok {
				return fmt.Errorf("connection kind must be a string or an array of string")
			}
			if value == "" {
				return fmt.Errorf("connection kind cannot be empty")
			}
			*m = append(*m, strings.ToLower(value))
		}
	default:
		return fmt.Errorf("connection kind must be a string or an array of string")
	}
	return nil
}

func (u *ConnectionKindMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = ConnectionKindMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("connection kind cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
