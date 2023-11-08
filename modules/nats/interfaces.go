// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
)

func Load(ctx caddy.Context) (App, error) {
	unm, err := ctx.App("nats")
	if err != nil {
		return nil, err
	}
	app, ok := unm.(App)
	if !ok {
		return nil, errors.New("nats: failed to type assert module type to nats.App")
	}
	return app, nil
}

type App interface {
	beyond.App
	beyond.BeyondAppLoader
	AddNewTokenBasedAuthPolicy(account string) (string, error)
	GetOptions() *embedded.Options
	GetServer() (*server.Server, error)
	ReloadServer() error
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

type Matcher interface {
	Match(request *jwt.AuthorizationRequestClaims) bool
}

type Keystore interface {
	Get(account string) (string, error)
}
