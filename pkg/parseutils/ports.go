// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils

import (
	"fmt"
	"strconv"
)

// ParsePort parses a port number. It returns an error if the value
// is not a valid port number (0-65535).
func ParsePort(value string) (int, error) {
	t, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("cannot decode port: %w", err)
	}
	port, err := UInt16(t)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %w", err)
	}
	return int(port), nil
}
