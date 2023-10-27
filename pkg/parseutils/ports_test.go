// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/parseutils"
)

func TestParsePort(t *testing.T) {
	cases := []struct {
		input  string
		output int
		expect bool
	}{
		{"-1", 0, false},
		{"0", 0, true},
		{"1", 1, true},
		{"65535", 65535, true},
		{"65536", 0, false},
	}
	for _, case_ := range cases {
		t.Run(case_.input, func(t *testing.T) {
			actual, err := parseutils.ParsePort(case_.input)
			if case_.expect && err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if !case_.expect && err == nil {
				t.Fatalf("expected error, got %d", actual)
			}
		})
	}
}

func TestParsePortError(t *testing.T) {
	cases := []struct {
		input string
	}{
		{"-1"},
		{"65536"},
		{"abc"},
	}
	for _, case_ := range cases {
		t.Run(case_.input, func(t *testing.T) {
			_, err := parseutils.ParsePort(case_.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
