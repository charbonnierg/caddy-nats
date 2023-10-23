// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

// Context returns the caddy context for the secrets app.
func (a *App) Context() caddy.Context {
	return a.ctx
}

func (a *App) Logger() *zap.Logger {
	return a.logger
}

// Get returns the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) Get(key string) (string, error) {
	store, secretkey, err := a.getStoreAndKey(key)
	if err != nil {
		return "", err
	}
	return store.Get(secretkey)
}

// Set sets the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) Set(key string, value string) error {
	store, secretkey, err := a.getStoreAndKey(key)
	if err != nil {
		return err
	}
	return store.Set(secretkey, value)
}

// GetStore returns the store with the given name.
// It allows fetching and saving secrets without specifying the store name.
func (a *App) GetStore(name string) secrets.Store {
	return a.getStore(name)
}

func (a *App) DefaultStore() secrets.Store {
	return a.getStore(a.defaultStore)
}

// AddSecretsReplacerVars adds replacer variables to the given replacer.
// It is used by other caddy modules to add secrets replacer variables to their own replacer.
func (a *App) AddSecretsReplacerVars(repl *caddy.Replacer) {
	repl.Map(func(key string) (any, bool) {
		secretsPrefix := "secret."
		if !strings.HasPrefix(key, secretsPrefix) {
			return nil, false
		}
		secretKey := strings.TrimPrefix(key, secretsPrefix)
		value, err := a.Get(secretKey)
		if err != nil {
			return nil, true
		}
		return string(value), true
	})
}
