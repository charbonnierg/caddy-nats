// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"errors"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond/modules/secrets"
)

var (
	DEFAULT_STORE_NAME = "default"
)

// Get returns the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) Get(key string) ([]byte, error) {
	store, secretkey, err := a.getStoreAndKey(key)
	if err != nil {
		return nil, err
	}
	return store.Get(secretkey)
}

// Set sets the value of the secret with the given name located in the given store.
// The key syntax is "secretkey@storename".
// If storename is not specified, the default store is used.
func (a *App) Set(key string, value []byte) error {
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

func (a *App) getStore(name string) *Store {
	return nil
}

func (a *App) getStoreAndKey(key string) (*Store, string, error) {
	parts := strings.Split(key, "@")
	var storename string
	var secretkey string
	switch len(parts) {
	case 1:
		storename = DEFAULT_STORE_NAME
		secretkey = parts[0]
	case 2:
		storename = parts[0]
		secretkey = parts[1]
	default:
		return nil, "", errors.New("invalid key")
	}
	store := a.getStore(storename)
	if store == nil {
		return nil, "", errors.New("store not found")
	}
	return store, secretkey, nil
}
