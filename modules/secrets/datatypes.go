// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package secrets

// Stores is a map of stores.
type Stores = map[string]Store

// Sources is a list of sources.
type Sources = []*Source

// Secrets is a list of secrets.
type Secrets = []*Secret

// Secret is used to pass secret values to functions.
type Secret struct {
	Source *Source
	Value  string
}

// Source is used to retrieve a secret from a store.
type Source struct {
	Store     Store
	StoreName string
	Key       string
}

// String returns the string representation of the source.
func (s *Source) String() string {
	return s.Key + "@" + s.StoreName
}

// Get returns the value of the secret.
func (s *Source) Get() (string, error) {
	return s.Store.Get(s.Key)
}
