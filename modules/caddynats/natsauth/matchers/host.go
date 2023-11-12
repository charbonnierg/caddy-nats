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

func (HostMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.host",
		New: func() caddy.Module { return new(HostMatcher) },
	}
}

type HostMatcher []string

func (m *HostMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *HostMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, host := range *m {
		switch host {
		case "in_process":
			if claims.ClientInformation.Host == "" {
				return true, nil
			}
		default:
			if host == claims.ClientInformation.Host {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *HostMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = HostMatcher{v}
	case []interface{}:
		*m = HostMatcher{}
		for _, value := range v {
			value, ok := value.(string)
			if !ok {
				return fmt.Errorf("host must be a string or an array of string")
			}
			if value == "" {
				return fmt.Errorf("host cannot be empty")
			}
			*m = append(*m, value)
		}
	default:
		return fmt.Errorf("host must be a string or an array of string")
	}
	return nil
}

func (u *HostMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = HostMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("host cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
