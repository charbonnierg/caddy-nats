// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package secrets provides a simple interface for managing secrets within caddy modules.
package oauth2

import (
	"errors"
)

var (
	// ErrInvalidSecretsModule is returned when the secrets module is invalid.
	ErrInvalidOauth2Module = errors.New("invalid oauth2 module")
	// ErrInvalidStoreModule is returned when a store module is invalid.
	ErrInvalidStoreModule = errors.New("invalid oauth2 store module")
)
