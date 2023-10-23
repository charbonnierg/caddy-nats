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
	"strconv"
)

// ParsePort parses a port number. It returns an error if the value
// is not a valid port number (0-65535).
func ParsePort(value string) (int, error) {
	t, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("cannot decode port: %v", err)
	}
	if t < 0 || t > 65535 {
		return 0, fmt.Errorf("cannot decode port: %d", t)
	}
	return t, nil
}
