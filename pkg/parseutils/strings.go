// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils

import "strconv"

// ParseInt32 converts string to int32 in a safe way.
// You get error when the value is out of the 32-bit range.
//
// This is a wrapper function of strconv.ParseInt.
func ParseInt32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

// ParseUInt32 converts string to uint32 in a safe way.
// You get error when the value is out of the 32-bit unsigned range.
//
// This is a wrapper function of strconv.ParseUint.
func ParseUInt32(s string) (uint32, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

// ParseInt16 converts string to int16 in a safe way.
// You get error when the value is out of the 16-bit range.
//
// This is a wrapper function of strconv.ParseInt.
func ParseInt16(s string) (int16, error) {
	i, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(i), nil
}

// ParseUInt16 converts string to uint16 in a safe way.
// You get error when the value is out of the 16-bit unsigned range.
//
// This is a wrapper function of strconv.ParseUint.
func ParseUInt16(s string) (uint16, error) {
	i, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(i), nil
}

// ParseInt8 converts string to int8 in a safe way.
// You get error when the value is out of the 8-bit range.
//
// This is a wrapper function of strconv.ParseInt.
func ParseInt8(s string) (int8, error) {
	i, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(i), nil
}

// ParseUInt8 converts string to uint8 in a safe way.
// You get error when the value is out of the 8-bit unsigned range.
//
// This is a wrapper function of strconv.ParseUint.
func ParseUInt8(s string) (uint8, error) {
	i, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(i), nil
}
