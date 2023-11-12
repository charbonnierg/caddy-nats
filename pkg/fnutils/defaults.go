// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package fnutils

func DefaultIfNil[T any](value *T, defaultValue *T) *T {
	if value == nil {
		return defaultValue
	}
	return value
}

func DefaultIfEmpty[T any](value []T, defaultValue []T) []T {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func DefaultIfEmptyMap[K comparable, V any](value map[K]V, defaultValue map[K]V) map[K]V {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func DefaultIfEmptyString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
