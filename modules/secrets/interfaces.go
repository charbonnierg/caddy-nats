package secrets

import (
	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

// SecretApp is the interface implemented by the secrets caddy app module.
type SecretApp interface {
	// Context returns the caddy context for the secrets app.
	Context() caddy.Context
	// Logger returns the logger for the secrets app.
	Logger() *zap.Logger
	// GetStore returns the store with the given name.
	GetStore(store string) Store
	// Get returns the value of the secret with the given name located in the given store.
	Get(key string) (string, error)
	// Set sets the value of the secret with the given name located in the given store.
	Set(key string, value string) error
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
	// Set sets the value of the secret with the given name.
	Set(key string, value string) error
}
