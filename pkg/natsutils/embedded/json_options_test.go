// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package embedded_test

import (
	"testing"

	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
)

func TestDefaultOptions(t *testing.T) {
	opts := embedded.NewOptions()
	if opts == nil {
		t.Fatal("Expected options to be created")
	}
	serverOpts, err := opts.GetServerOptions()
	if err != nil {
		t.Fatal(err)
	}
	if serverOpts == nil {
		t.Fatal("Expected nats options to be created")
	}
}
