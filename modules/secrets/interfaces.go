// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
)

// App is the interface implemented by the secrets caddy app module.
type App interface {
	beyond.App
	// GetStore returns the store with the given name.
	GetStore(store string) Store
	// GetSource returns the source with the given name.
	GetSource(name string) (*Source, error)
	// GetSecret returns the value of the secret with the given name located in the given store.
	GetSecret(key string) (*Secret, error)
	// AddAutomation adds an automation to the app and provisions it.
	AddAutomation(automation ...Automation) error
	// AddSecretsReplacerVars adds replacer variables to the given replacer.
	// It is used by other caddy modules to add secrets replacer variables to their own replacer.
	// It makes it possible to use secrets replacer variables in other caddy modules.
	AddSecretsReplacerVars(repl *caddy.Replacer)
}

// Store is a secret store.
// It is used to retrieve the value of a secret.
type Store interface {
	caddy.Module
	// Provision prepares the store for use.
	Provision(app App) error
	// Get returns the value of the secret with the given name.
	Get(key string) (string, error)
}

type Automation interface {
	Provision(app App) error
	Start() error
	Stop() error
}

// Trigger is a secret automation trigger.
// It is used to trigger the automation.
type Trigger interface {
	caddy.Module
	Provision(app App, automation Automation) error
	Subscribe(ctx context.Context) <-chan context.Context
}

// Template is a secret automation template.
// It is used to render the secret values into a string.
type Template interface {
	caddy.Module
	Provision(app App, automation Automation) error
	Render(input Secrets) (string, error)
}

// Handler is a secret automation handler.
// It is used to handle the secret value when it is fetched.
type Handler interface {
	caddy.Module
	Provision(app App, automation Automation) error
	Handle(value string) (string, error)
}
