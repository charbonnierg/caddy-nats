// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package datatypes

type Environ map[string]string

func (e Environ) Get(key string, default_ string) string {
	value, ok := e[key]
	if !ok {
		return default_
	}
	return value
}

func (e Environ) Set(key string, value string) {
	e[key] = value
}

func (e Environ) Entries() []string {
	entries := make([]string, len(e))
	i := 0
	for k, v := range e {
		entries[i] = k + "=" + v
		i++
	}
	return entries
}

func (e Environ) Keys() []string {
	keys := make([]string, len(e))
	i := 0
	for k := range e {
		keys[i] = k
		i++
	}
	return keys
}
