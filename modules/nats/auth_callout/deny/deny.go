// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package deny

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats/natsapp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(DenyAuthCallout{})
}

// A minimal auth callout handler that always denies access.
type DenyAuthCallout struct {
	logger  *zap.Logger
	err     error
	Message string `json:"message,omitempty"`
}

func (DenyAuthCallout) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.auth_callout.deny",
		New: func() caddy.Module { return new(DenyAuthCallout) },
	}
}

func (a *DenyAuthCallout) Provision(app *natsapp.App) error {
	a.logger = app.Context().Logger().Named("deny")
	if a.Message == "" {
		a.err = errors.New("access denied")
	} else {
		a.err = errors.New(a.Message)
	}
	return nil
}

func (a *DenyAuthCallout) Handle(request *natsapp.AuthorizationRequest) (*jwt.UserClaims, error) {
	a.logger.Info("denying access", zap.Any("client_infos", request.Claims.ClientInformation))
	return nil, a.err
}

var (
	_ natsapp.AuthCallout = (*DenyAuthCallout)(nil)
)
