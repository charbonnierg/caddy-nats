// Copyright 2023 Guillaume Charbonnier
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/utils"
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
			actual, err := utils.ParseBytes(case_.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != case_.expected {
				t.Errorf("expected %d, got %d", case_.expected, actual)
			}
		})
	}
}
