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
			actual, err := utils.ParsePort(case_.input)
			if case_.expect && err != nil {
				t.Fatalf("unexpected error: %v", err)
			} else if !case_.expect && err == nil {
				t.Fatalf("expected error, got %d", actual)
			}
		})
	}
}
