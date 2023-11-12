// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (ProtocolMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.protocol",
		New: func() caddy.Module { return new(ProtocolMatcher) },
	}
}

type ProtocolMatcher []int

func (m *ProtocolMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m ProtocolMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, protocol := range m {
		if protocol == claims.ConnectOptions.Protocol {
			return true, nil
		}
	}
	return false, nil
}

func (m *ProtocolMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case int:
		*m = ProtocolMatcher{v}
	case []interface{}:
		*m = ProtocolMatcher{}
		for _, value := range v {
			value, ok := value.(float64)
			if float64(int(value)) != value {
				return fmt.Errorf("each protocol must be an integer: bad value: %v", value)
			}
			if !ok {
				return fmt.Errorf("each protocol must be an integer: bad value: %v", value)
			}
			*m = append(*m, int(value))
		}
	default:
		return fmt.Errorf("protocol must be an integer or an array of integers: %v", v)
	}
	return nil
}

func (u *ProtocolMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = ProtocolMatcher{}
	for d.NextArg() {
		var proto int
		if err := parser.ParseInt(d, &proto, parser.Inplace()); err != nil {
			return err
		}
		*u = append(*u, proto)
	}
	return nil
}
