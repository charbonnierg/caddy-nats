// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package datatypes

func NewMap[T comparable, U any]() Map[T, U] {
	return Map[T, U]{}
}

type Map[T comparable, U any] map[T]U

func (m Map[T, U]) Keys() []T {
	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func (m Map[T, U]) Values() []U {
	values := make([]U, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func (m Map[T, U]) Pop(key T) (U, bool) {
	value, ok := m[key]
	if ok {
		delete(m, key)
	}
	return value, ok
}
