// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secretsapp

import (
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
)

func (a *App) getStore(name string) secrets.Store {
	return a.stores[name]
}

func (a *App) get(key string) (string, error) {
	source, err := a.getSource(key)
	if err != nil {
		return "", err
	}
	return source.Store.Get(source.Key)
}

func (a *App) getSource(key string) (*secrets.Source, error) {
	parts := strings.Split(key, "@")
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
	store := a.getStore(storename)
	if store == nil {
		return nil, secrets.ErrStoreNotFound
	}
	return &secrets.Source{
		StoreName: storename,
		Store:     store,
		Key:       secretkey,
	}, nil
}

func (a *App) addReplacerVars(repl *caddy.Replacer) {
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
