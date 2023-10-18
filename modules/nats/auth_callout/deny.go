// SPDX-License-Identifier: Apache-2.0

package auth_callout

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond/modules/nats/natsapp"
	"github.com/nats-io/jwt/v2"
)

func init() {
	caddy.RegisterModule(DenyAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type DenyAuthCallout struct{}

func (DenyAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.deny",
		New: func() caddy.Module { return new(DenyAuthCallout) },
	}
}

func (a *DenyAuthCallout) Provision(app *natsapp.App) error {
	return nil
}

func (a *DenyAuthCallout) Handle(request *natsapp.AuthorizationRequest) (*jwt.UserClaims, error) {
	return nil, errors.New("access denied")
}

var (
	_ natsapp.AuthCallout = (*DenyAuthCallout)(nil)
)
