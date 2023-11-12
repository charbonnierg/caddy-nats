// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2

import "github.com/caddyserver/caddy/v2"

const (
	// Namespace used by the oauth2 module.
	NS = "oauth2"
	// Namespaces used by the oauth2 stores modules.
	STORES_NS = "oauth2.stores"
)

// StoreID returns the store ID for the given name.
func StoreID(name string) caddy.ModuleID {
	return caddy.ModuleID(STORES_NS + "." + name)
}
