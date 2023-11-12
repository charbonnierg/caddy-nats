package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func (TokenMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.matchers.token",
		New: func() caddy.Module { return new(TokenMatcher) },
	}
}

type TokenMatcher []string

func (m *TokenMatcher) Provision(ctx caddy.Context) error {
	return nil
}

func (m *TokenMatcher) Match(request natsauth.AuthorizationRequest) (bool, error) {
	claims := request.Claims()
	for _, token := range *m {
		if token == claims.ConnectOptions.Token {
			return true, nil
		}
	}
	return false, nil
}

func (m *TokenMatcher) UnmarshalJSON(b []byte) error {
	var val interface{}
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case string:
		*m = TokenMatcher{v}
	case []interface{}:
		*m = TokenMatcher{}
		for _, value := range v {
			value, ok := value.(string)
			if !ok {
				return fmt.Errorf("token must be a string or an array of string")
			}
			if value == "" {
				return fmt.Errorf("token cannot be empty")
			}
			*m = append(*m, value)
		}
	default:
		return fmt.Errorf("token must be a string or an array of string")
	}
	return nil
}

func (u *TokenMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	*u = TokenMatcher{}
	for d.NextArg() {
		val := d.Val()
		if val == "" {
			return d.Err("token cannot be empty")
		}
		*u = append(*u, d.Val())
	}
	return nil
}
