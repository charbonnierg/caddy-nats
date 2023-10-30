// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parseutils

import (
	"math"
)

// Int32 converts int to int32 in a safe way.
// You get error when the value is out of the 32-bit range (-2147483648 through 2147483647).
func Int32(i int) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, ErrInt32OutOfRange
	}
	return int32(i), nil
}

// UInt32 converts int to uint32 in a safe way.
// You get error when the value is out of the 32-bit unsigned range (0 through 4294967295).
func UInt32(i int) (uint32, error) {
	if i > (1<<32-1) || i < 0 {
		return 0, ErrUInt32OutOfRange
	}
	return uint32(i), nil
}

// Int16 converts int to int16 in a safe way.
// You get error when the value is out of the 16-bit range (-32768 through 32767).
func Int16(i int) (int16, error) {
	if i > math.MaxInt16 || i < math.MinInt16 {
		return 0, ErrInt16OutOfRange
	}
	return int16(i), nil
}

// UInt16 converts int to uint16 in a safe way.
// You get error when the value is out of the 16-bit unsigned range (0 through 65535).
func UInt16(i int) (uint16, error) {
	if i > math.MaxUint16 || i < 0 {
		return 0, ErrUInt16OutOfRange
	}
	return uint16(i), nil
}

// Int8 converts int to int8 in a safe way.
// You get error when the value is out of the 8-bit range (-128 through 127).
func Int8(i int) (int8, error) {
	if i > math.MaxInt8 || i < math.MinInt8 {
		return 0, ErrInt8OutOfRange
	}
	return int8(i), nil
}

// UInt8 converts int to uint8 in a safe way.
// You get error when the value is out of the 8-bit unsigned range (0 through 255).
func UInt8(i int) (uint8, error) {
	if i > math.MaxUint8 || i < 0 {
		return 0, ErrUInt8OutOfRange
	}
	return uint8(i), nil
}
