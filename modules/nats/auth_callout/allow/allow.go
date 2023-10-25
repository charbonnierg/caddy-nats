// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package allow

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats/natsapp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AllowAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type AllowAuthCallout struct {
	logger   *zap.Logger
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
	c.logger = app.Context().Logger().Named("allow")
	return nil
}

func (a *AllowAuthCallout) Handle(request *natsapp.AuthorizationRequest) (*jwt.UserClaims, error) {
	if request.Claims == nil {
		// If the request is not authenticated, deny access
		return nil, errors.New("not authenticated")
	}
	userClaims := jwt.NewUserClaims(request.Claims.UserNkey)
	if a.Template != nil {
		a.logger.Info("rendering template", zap.Any("template", a.Template))
		// Apply the template
		a.Template.Render(request, userClaims)
	}
	userClaims.Audience = request.ReplaceAll(a.Account, "")
	if userClaims.Audience == "" {
		// If the target account is still empty, deny access
		return nil, errors.New("no target account specified")
	}
	a.logger.Info("allowing access", zap.Any("user_claims", userClaims))
	// And that's it, return the user claims
	return userClaims, nil
}

var (
	_ natsapp.AuthCallout = (*AllowAuthCallout)(nil)
)
