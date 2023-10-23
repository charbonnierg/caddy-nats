// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond/modules/secrets"
)

// MemoryStore is a secrets store that stores secrets in memory.
// It is not persistent and is only useful for testing or sharing
// non-persistent secrets between caddy modules when combined with
// random replacer variables.
// An empty MemoryStore is not usable, it must be provisioned first.
type MemoryStore struct {
	secrets map[string]string
}

// Provision prepares the store for use.
func (s *MemoryStore) Provision(app secrets.SecretApp) error {
	s.secrets = make(map[string]string)
	return nil
}

func (s *MemoryStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.store.memory",
		New: func() caddy.Module { return new(MemoryStore) },
	}
}

// Get retrieves a value from the store for a given key.
func (s *MemoryStore) Get(key string) (string, error) {
	value, ok := s.secrets[key]
	if !ok {
		return "", secrets.ErrSecretNotFound
	}
	return value, nil
}

// Set writes a value to the store for a given existing key.
func (s *MemoryStore) Set(key string, value string) error {
	s.secrets[key] = value
	return nil
}

// Interface guards
var (
	_ secrets.Store = (*MemoryStore)(nil)
)
