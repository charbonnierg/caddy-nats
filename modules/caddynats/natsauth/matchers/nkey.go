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

func (NKeyMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.nkey",
		New: func() caddy.Module { return new(NKeyMatcher) },
	}
}

type NKeyMatcher []string

func (m *NKeyMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *NKeyMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, nkey := range *m {
		if nkey == claims.ConnectOptions.Nkey {
			return true, nil
		}
	}
	return false, nil
}

func (m *NKeyMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = NKeyMatcher{v}
	case []interface{}:
		*m = NKeyMatcher{}
		for _, value := range v {
			value, ok := value.(string)
			if !ok {
				return fmt.Errorf("nkey must be a string or an array of string")
			}
			if value == "" {
				return fmt.Errorf("nkey cannot be empty")
			}
			*m = append(*m, value)
		}
	default:
		return fmt.Errorf("nkey must be a string or an array of string")
	}
	return nil
}

func (u *NKeyMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = NKeyMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("nkey cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
