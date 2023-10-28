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
func (a *App) Context() caddy.Context { return a.ctx }

// Logger returns the logger for the secrets app.
func (a *App) Logger() *zap.Logger { return a.logger }

// GetStore returns the store with the given name.
func (a *App) GetStore(name string) secrets.Store { return a.stores[name] }

// GetDefaultStore returns the default store.
func (a *App) GetDefaultStore() secrets.Store { return a.stores[a.defaultStore] }

// GetSecret returns the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) GetSecret(key string) (*secrets.Secret, error) {
	source, err := a.GetSource(key)
	if err != nil {
		return nil, err
	}
	value, err := source.Get()
	if err != nil {
		return nil, err
	}
	return &secrets.Secret{
		Source: source,
		Value:  value,
	}, nil
}

// GetSource returns the source with the given name.
func (a *App) GetSource(name string) (*secrets.Source, error) {
	parts := strings.Split(name, "@")
	var storename string
	var secretkey string
	switch len(parts) {
	case 1:
		// If there is only one part, it is the secret key
		secretkey = parts[0]
		storename = a.defaultStore
	case 2:
		// If there are two parts, the first one is the secret key and the second one is the store name
		secretkey = parts[0]
		storename = parts[1]
	default:
		// if there are more than 2 parts, we cannot be sure of the store name
		// so we return an error
		return nil, secrets.ErrInvalidKey
	}
	store := a.stores[storename]
	if store == nil {
		return nil, secrets.ErrStoreNotFound
	}
	return &secrets.Source{
		StoreName: storename,
		Store:     store,
		Key:       secretkey,
	}, nil
}

// AddAutomation adds an automation to the app and provisions it.
func (a *App) AddAutomation(automation ...secrets.Automation) error {
	for _, auto := range automation {
		if err := auto.Provision(a); err != nil {
			return err
		}
		a.automations = append(a.automations, auto)
	}
	return nil
}

// AddStore adds a store to the app and provisions it.
func (a *App) AddStore(name string, store secrets.Store) error {
	if err := store.Provision(a); err != nil {
		return err
	}
	a.stores[name] = store
	// Set default store if it is not already set
	if a.defaultStore == "" {
		a.defaultStore = name
	}
	return nil
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
		value, err := a.GetSecret(secretKey)
		if err != nil {
			return nil, true
		}
		return string(value.Value), true
	})
}
