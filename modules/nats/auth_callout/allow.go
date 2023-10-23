// SPDX-License-Identifier: Apache-2.0

package auth_callout

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats/natsapp"
)

func init() {
	caddy.RegisterModule(AllowAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type AllowAuthCallout struct {
	User     string            `json:"user,omitempty"`
	Account  string            `json:"account,omitempty"`
	Template *natsapp.Template `json:"template,omitempty"`
}

func (AllowAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.allow",
		New: func() caddy.Module { return new(AllowAuthCallout) },
	}
}

func (c *AllowAuthCallout) Provision(app *natsapp.App) error {
	return nil
}

func (a *AllowAuthCallout) Handle(request *natsapp.AuthorizationRequest) (*jwt.UserClaims, error) {
	userClaims := jwt.NewUserClaims(request.Claims.UserNkey)
	if a.Template != nil {
		// Apply the template
		a.Template.Render(request, userClaims)
	}
	if a.Account != "" {
		// The target account must be specified as JWT audience
		userClaims.Audience = request.ReplaceAll(a.Account, "")
	} else {
		// If not specified, the target account is the username
		userClaims.Audience = request.Claims.ConnectOptions.Username
	}
	if userClaims.Audience == "" {
		// If the target account is still empty, deny access
		return nil, errors.New("no target account specified")
	}
	// And that's it, return the user claims
	return userClaims, nil
}

var (
	_ natsapp.AuthCallout = (*AllowAuthCallout)(nil)
)
