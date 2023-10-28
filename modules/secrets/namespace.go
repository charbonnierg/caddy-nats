// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package secrets provides a simple interface for managing secrets within caddy modules.
package secrets

import (
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
