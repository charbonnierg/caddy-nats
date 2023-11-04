// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package datatypes

// NewSet creates a new set
func NewSet[T comparable](value ...T) Set[T] {
	s := Set[T]{}
	if len(value) > 0 {
		s.Add(value[0], value[1:]...)
	}
	return s
}

// Set is a generic set of unique values
type Set[T comparable] map[T]struct{}

// StringSet is a set of strings
type StringSet = Set[string]

// ByteSet is a set of bytes
type ByteSet = Set[byte]

// Add adds one or several subjects to the set
func (s Set[T]) Add(value T, values ...T) {
	s[value] = struct{}{}
	for _, sub := range values {
		s[sub] = struct{}{}
	}
}

// Remove removes one or several subjects from the set
func (s Set[T]) Remove(value ...T) {
	for _, sub := range value {
		delete(s, sub)
	}
}

// Slice returns the list of subjects in the set
func (s Set[T]) Slice() []T {
	subjects := make([]T, 0, len(s))
	for subject := range s {
		subjects = append(subjects, subject)
	}
	return subjects
}

// Contains returns true if the set contains the given subject
func (s Set[T]) Contains(value T) bool {
	_, ok := s[value]
	return ok
}

// IsEmpty returns true if the set is empty
func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

// IsEqual returns true if the set is equal to the given set
func (s Set[T]) IsEqual(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for subject := range s {
		if !other.Contains(subject) {
			return false
		}
	}
	return true
}
