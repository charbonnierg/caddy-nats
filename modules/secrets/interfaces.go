// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/pkg/datatypes"
)

// SecretApp is the interface implemented by the secrets caddy app module.
type SecretApp interface {
	caddy.Module
	caddy.App
	// Context returns the caddy context for the secrets app.
	Context() caddy.Context
	// GetStore returns the store with the given name.
	GetStore(store string) Store
	// GetSource returns the source with the given name.
	GetSource(name string) (*Source, error)
	// Get returns the value of the secret with the given name located in the given store.
	Get(key string) (string, error)
	// AddSecretsReplacerVars adds replacer variables to the given replacer.
	// It is used by other caddy modules to add secrets replacer variables to their own replacer.
	// It makes it possible to use secrets replacer variables in other caddy modules.
	AddSecretsReplacerVars(repl *caddy.Replacer)
}

type Store interface {
	caddy.Module
	// Provision prepares the store for use.
	Provision(app SecretApp) error
	// Get returns the value of the secret with the given name.
	Get(key string) (string, error)
}

type Source struct {
	Store     Store
	StoreName string
	Key       string
}

func (s Source) String() string {
	return fmt.Sprintf("%s@%s", s.Key, s.StoreName)
}

type Stores = datatypes.Map[string, Store]
