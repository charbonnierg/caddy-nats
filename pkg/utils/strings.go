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
