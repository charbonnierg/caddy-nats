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
	"math"
)

// Int32 converts int to int32 in a safe way.
// You get error when the value is out of the 32-bit range (-2147483648 through 2147483647).
func Int32(i int) (int32, error) {
	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, fmt.Errorf("int32 out of range: %d", i)
	}
	return int32(i), nil
}

// Int16 converts int to int16 in a safe way.
// You get error when the value is out of the 16-bit range (-32768 through 32767).
func Int16(i int) (int16, error) {
	if i > math.MaxInt16 || i < math.MinInt16 {
		return 0, fmt.Errorf("int16 out of range: %d", i)
	}
	return int16(i), nil
}

// Int16 converts int to uint16 in a safe way.
// You get error when the value is out of the 16-bit unsigned range (0 through 65535).
func UInt16(i int) (uint16, error) {
	if i > math.MaxUint16 || i < 0 {
		return 0, fmt.Errorf("uint16 out of range: %d", i)
	}
	return uint16(i), nil
}

// Int8 converts int to int8 in a safe way.
// You get error when the value is out of the 8-bit range (-128 through 127).
func Int8(i int) (int8, error) {
	if i > math.MaxInt8 || i < math.MinInt8 {
		return 0, fmt.Errorf("int8 out of range: %d", i)
	}
	return int8(i), nil
}

// UInt8 converts int to uint8 in a safe way.
// You get error when the value is out of the 8-bit unsigned range (0 through 255).
func UInt8(i int) (uint8, error) {
	if i > math.MaxUint8 || i < 0 {
		return 0, fmt.Errorf("uint8 out of range: %d", i)
	}
	return uint8(i), nil
}
