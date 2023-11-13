// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package secrets provides a simple interface for managing secrets within caddy modules.
package secrets

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
)

// Namespace used by the secrets module.
const (
	NS = "secrets"
	// Namespaces used by the secret stores modules.
	STORES_NS = "secrets.stores"
	// Namespaces used by the secret triggers modules.
	TRIGGER_NS = "secrets.triggers"
	// Namespaces used by the secret handlers modules.
	HANDLER_NS = "secrets.handlers"
)

var (
	// ErrInvalidSecretsModule is returned when the secrets module is invalid.
	ErrInvalidSecretsModule = errors.New("invalid secrets module")
	// ErrInvalidStoreModule is returned when a store module is invalid.
	ErrInvalidStoreModule = errors.New("invalid store module")
	// ErrInvalidAutomationModule is returned when an automation module is invalid.
	ErrInvalidAutomationModule = errors.New("invalid automation module")
	// ErrStoreNotFound is returned when a store is not found.
	ErrStoreNotFound = errors.New("store not found")
	// ErrSecretNotFound is returned when a secret is not found.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrSecretAlreadyExists is returned when a secret already exists.
	ErrSecretAlreadyExists = errors.New("secret already exists")
	// ErrInvalidKey is returned when a key is invalid.
	ErrInvalidKey = errors.New("invalid key")
)

// StoreID returns the store ID for the given name.
func StoreID(name string) caddy.ModuleID {
	return caddy.ModuleID(STORES_NS + "." + name)
}

// TriggerID returns the trigger ID for the given name.
func TriggerID(name string) caddy.ModuleID {
	return caddy.ModuleID(TRIGGER_NS + "." + name)
}

// HandlerID returns the handler ID for the given name.
func HandlerID(name string) caddy.ModuleID {
	return caddy.ModuleID(HANDLER_NS + "." + name)
}

// UpdateReplacer is used to update the replacer with the secrets.
func UpdateReplacer(ctx caddy.Context, repl *caddy.Replacer) error {
	unm, err := ctx.App(NS)
	if err != nil {
		return err
	}
	app, ok := unm.(App)
	if !ok {
		return ErrInvalidSecretsModule
	}
	app.UpdateReplacer(repl)
	return nil
}
