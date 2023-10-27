// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddyutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCaddyutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Caddyutils Suite")
}
