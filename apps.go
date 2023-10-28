// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package beyond

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// Apps is a map of caddy app modules registered in other namespaces
// than beyond namespaces.
type Apps map[caddy.ModuleID]App

// App is an extension to caddy.App interface that also requires the
// caddy.Module interface.
type App interface {
	caddy.Module
	caddy.App
	caddy.Provisioner
	caddy.Validator
	// Context returns the caddy context for the secrets app.
	Context() caddy.Context
	Logger() *zap.Logger
}

// Get returns the app with the given id (each app has a unique id)
func (a Apps) Get(id string) (App, bool) {
	mid := caddy.ModuleID(id)
	app, ok := a[mid]
	if !ok {
		return nil, false
	}
	return app, true
}

// Add adds the given app to the map.
// If an app with the same id already exists, an error is returned.
func (a Apps) Add(app App) error {
	moduleID := app.CaddyModule().ID
	if _, ok := a[moduleID]; ok {
		return fmt.Errorf("app already exists: %s", moduleID)
	}
	a[moduleID] = app
	return nil
}
