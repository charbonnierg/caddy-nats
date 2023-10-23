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

package utils

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// ParseBytes parses a human-readable byte value
// (e.g. "1M", "1MB", "1 MiB", "1 mebibyte", etc.)
// into an integer.
func ParseBytes(value string) (int, error) {
	switch value {
	// humanize.ParseBytes expects a positive value,
	// so we must handle -1 as a special case.
	// Other negative values are not supported.
	case "":
		return 0, nil
	case "-1":
		return -1, nil
	default:
		v, err := humanize.ParseBytes(value)
		if err != nil {
			return 0, fmt.Errorf("cannot decode size %s to int: %s", value, err.Error())
		}
		return int(v), nil
	}
}
