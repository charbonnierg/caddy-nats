// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

// Context returns the caddy context for the secrets app.
func (a *App) Context() caddy.Context { return a.ctx }

// Logger returns the logger for the secrets app.
func (a *App) Logger() *zap.Logger { return a.logger }

// GetStore returns the store with the given name.
func (a *App) GetStore(name string) secrets.Store { return a.getStore(name) }

// Get returns the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) Get(key string) (string, error) { return a.get(key) }

// GetSource returns the source with the given name.
func (a *App) GetSource(name string) (*secrets.Source, error) { return a.getSource(name) }

// DefaultStore returns the default store.
func (a *App) DefaultStore() secrets.Store { return a.getStore(a.defaultStore) }

// AddSecretsReplacerVars adds replacer variables to the given replacer.
// It is used by other caddy modules to add secrets replacer variables to their own replacer.
func (a *App) AddSecretsReplacerVars(repl *caddy.Replacer) { a.addReplacerVars(repl) }
