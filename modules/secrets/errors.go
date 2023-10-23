// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secrets

import "errors"

var (
	ErrSecretNotFound      = errors.New("secret not found")
	ErrSecretAlreadyExists = errors.New("secret already exists")
)
