// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils

import (
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
		return 0, ErrEmptyBytesSize
	case "-1":
		return -1, nil
	default:
		v, err := humanize.ParseBytes(value)
		if err != nil {
			return 0, ErrInvalidBytesSize
		}
		return int(v), nil
	}
}
