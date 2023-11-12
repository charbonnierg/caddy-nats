package matchers

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
)

func init() {
	caddy.RegisterModule(UsernameMatcher{})
	caddy.RegisterModule(PasswordMatcher{})
	caddy.RegisterModule(TokenMatcher{})
	caddy.RegisterModule(NKeyMatcher{})
	caddy.RegisterModule(VersionMatcher{})
	caddy.RegisterModule(ProtocolMatcher{})
	caddy.RegisterModule(NameMatcher{})
	caddy.RegisterModule(HostMatcher{})
	caddy.RegisterModule(ConnectionKindMatcher{})
	caddy.RegisterModule(ConnectionTypeMatcher{})
}

var (
	_ natsauth.AuthorizationMatcher = (*UsernameMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*PasswordMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*TokenMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*NKeyMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*VersionMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*ProtocolMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*NameMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*HostMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*ConnectionKindMatcher)(nil)
	_ natsauth.AuthorizationMatcher = (*ConnectionTypeMatcher)(nil)
)
