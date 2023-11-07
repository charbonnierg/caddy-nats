// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
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
	Config() *natsutils.AuthServiceConfig
	Handle(request *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error)
}

type AuthRequest interface {
	Claims() *jwt.AuthorizationRequestClaims
	Context() context.Context
	Replacer() *caddy.Replacer
}

type AuthCallout interface {
	Provision(app App) error
	SetAccount(account string) error
	Handle(claims AuthRequest) (*jwt.UserClaims, error)
}

type Template interface {
	caddyfile.Unmarshaler
	Render(request AuthRequest, user *jwt.UserClaims)
}

// InputConnector is a Caddy module that serves as a connector
// to a data source. It reads data from a data source and sends it
// to a stream.
type Connector interface {
	caddy.Module
	Provision(app App) error
	Start() error
	Stop() error
}
