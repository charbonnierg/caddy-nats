// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// This test file is not written using Ginkgo, but it is written using the
// standard testing package. This is because the Ginkgo test runner requires
// more boilerplate to run test tables, and the standard testing package
// provides a simpler way to do this.
package parseutils_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/parseutils"
)

func TestParseBytes(t *testing.T) {
	cases := []struct {
		input    string
		expected int
	}{
		{"-1", -1},
		{"0", 0},
		{"1", 1},
		{"1B", 1},
		{"1 B", 1},
		{"1 b", 1},
		{"1KB", 1000},
		{"1 KB", 1000},
		{"1 kb", 1000},
		{"1KiB", 1024},
		{"1 KiB", 1024},
		{"1 kib", 1024},
		{"1MB", 1000 * 1000},
		{"1 MB", 1000 * 1000},
		{"1MiB", 1024 * 1024},
		{"1 MiB", 1024 * 1024},
		{"1mib", 1024 * 1024},
		{"1GB", 1000 * 1000 * 1000},
		{"1 GB", 1000 * 1000 * 1000},
		{"1GiB", 1024 * 1024 * 1024},
		{"1 GiB", 1024 * 1024 * 1024},
		{"1gib", 1024 * 1024 * 1024},
		{"1TB", 1000 * 1000 * 1000 * 1000},
		{"1 TB", 1000 * 1000 * 1000 * 1000},
		{"1TiB", 1024 * 1024 * 1024 * 1024},
		{"1 TiB", 1024 * 1024 * 1024 * 1024},
		{"1tib", 1024 * 1024 * 1024 * 1024},
	}
	for _, case_ := range cases {
		t.Run(case_.input, func(t *testing.T) {
			actual, err := parseutils.ParseBytes(case_.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != case_.expected {
				t.Errorf("expected %d, got %d", case_.expected, actual)
			}
		})
	}
}

func TestParseBytesError(t *testing.T) {
	cases := []string{
		"",
		"abc",
	}
	for _, case_ := range cases {
		t.Run(case_, func(t *testing.T) {
			_, err := parseutils.ParseBytes(case_)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
