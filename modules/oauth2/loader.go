// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
)

// LoadCaddyApp can be used to load the secrets app using caddy context.
// Use this method when your module has unidirectional dependency on secrets module.
func LoadCaddyApp(ctx *caddy.Context) (App, error) {
	unm, err := ctx.App(NS)
	if err != nil {
		return nil, err
	}
	secrets, ok := unm.(App)
	if !ok {
		return nil, ErrInvalidOauth2Module
	}
	return secrets, nil
}

// LoadBeyondApp can be used to load the secrets app using beyond app.
// Use this method when your module has bidirectional dependency on secrets module.
func LoadBeyondApp(b *beyond.Beyond) (App, error) {
	app, err := b.LoadApp(NS)
	if err != nil {
		return nil, err
	}
	secretsApp, ok := app.(App)
	if !ok {
		return nil, ErrInvalidOauth2Module
	}
	return secretsApp, nil
}
