// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2app

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// generateRandomASCIIString returns a securely generated random ASCII string.
// It reads random numbers from crypto/rand and searches for printable characters.
// It will return an error if the system's secure random number generator fails to
// function correctly, in which case the caller must not continue.
func generateRandomASCIIString(length int8) (string, error) {
	if length < 32 {
		return "", fmt.Errorf("length must be at least 32")
	}
	if length > 127 {
		return "", fmt.Errorf("length must be at most 127")
	}
	var max = big.NewInt(int64(127))
	var size int8 = 0
	result := ""
	for {
		if size >= length {
			return result, nil
		}
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("error reading random number: %v", err)
		}
		n := uint8(num.Int64())
		// Make sure that the number/byte/letter is inside
		// the range of printable ASCII characters (excluding space and DEL)
		if n > 32 && n < 127 {
			result += string(n)
			size++
		}
	}
}
