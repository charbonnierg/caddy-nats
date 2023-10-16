// SPDX-License-Identifier: Apache-2.0

package auth_callout

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/modules"
	"github.com/nats-io/jwt/v2"
)

func init() {
	caddy.RegisterModule(DenyAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type DenyAuthCallout struct{}

func (DenyAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.always_deny",
		New: func() caddy.Module { return new(DenyAuthCallout) },
	}
}

func (a *DenyAuthCallout) Provision(app *modules.App) error {
	return nil
}

func (a *DenyAuthCallout) Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	return nil, errors.New("access denied")
}

var (
	_ modules.AuthCallout = (*DenyAuthCallout)(nil)
)
