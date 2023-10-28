// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package secrets provides a simple interface for managing secrets within caddy modules.
package secrets

import (
	"errors"
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
