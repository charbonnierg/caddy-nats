// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"github.com/caddyserver/caddy/v2"
)

func Load(ctx caddy.Context) (SecretApp, error) {
	unm, err := ctx.App(APP_NAME)
	if err != nil {
		return nil, err
	}
	secrets, ok := unm.(SecretApp)
	if !ok {
		return nil, ErrInvalidSecretsModule
	}
	return secrets, nil
}
