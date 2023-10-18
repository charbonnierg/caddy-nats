package beyond

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
)

type Apps map[caddy.ModuleID]App

type App interface {
	caddy.Module
	caddy.App
}

func (a Apps) Get(id string) (App, bool) {
	mid := caddy.ModuleID(id)
	app, ok := a[mid]
	if !ok {
		return nil, false
	}
	return app, true
}

func (a Apps) Add(app App) error {
	moduleID := app.CaddyModule().ID
	if _, ok := a[moduleID]; ok {
		return fmt.Errorf("app already exists: %s", moduleID)
	}
	a[moduleID] = app
	return nil
}
