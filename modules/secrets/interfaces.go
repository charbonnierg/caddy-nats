package secrets

import (
	"github.com/caddyserver/caddy/v2"
)

// SecretApp is the interface implemented by the secrets caddy app module.
type SecretApp interface {
	// GetStore returns the store with the given name.
	GetStore(store string) Store
	// Get returns the value of the secret with the given name located in the given store.
	Get(key string) ([]byte, error)
	// Set sets the value of the secret with the given name located in the given store.
	Set(key string, value []byte) error
	// AddSecretsReplacerVars adds replacer variables to the given replacer.
	// It is used by other caddy modules to add secrets replacer variables to their own replacer.
	// It makes it possible to use secrets replacer variables in other caddy modules.
	AddSecretsReplacerVars(repl *caddy.Replacer)
}

type Store interface {
	// Get returns the value of the secret with the given name.
	Get(key string) ([]byte, error)
	// Set sets the value of the secret with the given name.
	Set(key string, value []byte) error
}
