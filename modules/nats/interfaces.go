// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/nats/embedded"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

type App interface {
	beyond.App
	beyond.BeyondAppLoader
	Options() *embedded.Options
	GetAuthUserPass() (string, string, error)
	GetServer() (*server.Server, error)
	ReloadServer() error
}

type AuthService interface {
	Client() *natsutils.Client
	Provision(app App) error
	Listen(conn *nats.Conn) error
	Close() error
}

type AuthRequest interface {
	Claims() *jwt.AuthorizationRequestClaims
	Context() context.Context
	Replacer() *caddy.Replacer
}

type AuthCallout interface {
	Provision(app App) error
	Handle(request AuthRequest) (*jwt.UserClaims, error)
}

type Template interface {
	caddyfile.Unmarshaler
	Render(request AuthRequest, user *jwt.UserClaims)
}
