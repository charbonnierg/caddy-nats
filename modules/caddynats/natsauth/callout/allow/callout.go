// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package allow

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AllowAuthCallout{})
}

func (AllowAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.callouts.allow",
		New: func() caddy.Module { return new(AllowAuthCallout) },
	}
}

// A minimal auth callout handler that always denies access.
type AllowAuthCallout struct {
	account  string
	logger   *zap.Logger
	Template *natsauth.Template `json:"template,omitempty"`
}

func (c *AllowAuthCallout) Provision(ctx caddy.Context, account string) error {
	c.logger = ctx.Logger().Named("allow")
	c.account = account
	return nil
}

func (a *AllowAuthCallout) Handle(request natsauth.AuthorizationRequest) (*jwt.UserClaims, error) {
	claims := request.Claims()
	repl := request.Replacer()
	if claims == nil {
		// If the request is not authenticated, deny access
		return nil, errors.New("not authenticated")
	}
	userClaims := jwt.NewUserClaims(claims.UserNkey)
	if a.Template != nil {
		a.logger.Info("rendering template", zap.Any("template", a.Template))
		// Apply the template
		a.Template.Render(request, userClaims)
	}
	userClaims.Audience = repl.ReplaceAll(a.account, "")
	if userClaims.Audience == "" {
		// If the target account is still empty, deny access
		return nil, errors.New("no target account specified")
	}
	a.logger.Info("allowing access", zap.Any("jwt_claims", userClaims.ClaimsData), zap.Any("user_claims", userClaims.User))
	// And that's it, return the user claims
	return userClaims, nil
}
