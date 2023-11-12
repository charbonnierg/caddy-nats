// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

func (a *App) Logger() *zap.Logger {
	return a.logger
}

func (a *App) Context() caddy.Context {
	return a.ctx
}
