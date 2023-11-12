package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func (PasswordMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.password",
		New: func() caddy.Module { return new(PasswordMatcher) },
	}
}

type PasswordMatcher []string

func (m *PasswordMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *PasswordMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, password := range *m {
		if password == claims.ConnectOptions.Password {
			return true, nil
		}
	}
	return false, nil
}

func (m *PasswordMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = PasswordMatcher{v}
	case []interface{}:
		*m = PasswordMatcher{}
		for _, name := range v {
			name, ok := name.(string)
			if !ok {
				return fmt.Errorf("password must be a string or an array of string")
			}
			if name == "" {
				return fmt.Errorf("password cannot be empty")
			}
			*m = append(*m, name)
		}
	default:
		return fmt.Errorf("password must be a string or an array of string")
	}
	return nil
}

func (u *PasswordMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = PasswordMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("password cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
