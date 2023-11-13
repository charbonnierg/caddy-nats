// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/modules/python"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("python", parseGlobalOption)
}

type App struct {
	ctx       caddy.Context
	repl      *caddy.Replacer
	Processes []*PythonProcess `json:"processes,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "python",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Context() caddy.Context {
	return a.ctx
}

func (a *App) Logger() *zap.Logger {
	return a.ctx.Logger(a)
}

func (a *App) Replacer() *caddy.Replacer {
	return a.repl
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	repl := caddy.NewReplacer()
	if err := secrets.UpdateReplacer(ctx, repl); err != nil {
		return err
	}
	a.repl = repl
	for _, process := range a.Processes {
		if err := process.Provision(a); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Validate() error {
	for _, process := range a.Processes {
		if process.Name == "" {
			return errors.New("process name is required")
		}
		if process.Command == "" {
			return errors.New("process command is required")
		}
	}
	return nil
}

func (a *App) Start() error {
	for _, process := range a.Processes {
		if err := process.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Stop() error {
	for _, process := range a.Processes {
		if err := process.Stop(); err != nil {
			return err
		}
	}
	return nil
}

var (
	_ python.App = (*App)(nil)
)
