package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func (UsernameMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.username",
		New: func() caddy.Module { return new(UsernameMatcher) },
	}
}

type UsernameMatcher []string

func (m *UsernameMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m UsernameMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, username := range m {
		if username == claims.ConnectOptions.Username {
			return true, nil
		}
		if username == claims.ClientInformation.User {
			return true, nil
		}
	}
	return false, nil
}

func (m *UsernameMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = UsernameMatcher{v}
	case []interface{}:
		*m = UsernameMatcher{}
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

func (u *UsernameMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = UsernameMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("username cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
