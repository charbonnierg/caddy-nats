// SPDX-License-Identifier: Apache-2.0

package natsoptions_test

import (
	"testing"

	"github.com/charbonnierg/caddy-nats/embedded/natsoptions"
)

func TestDefaultOptions(t *testing.T) {
	opts := natsoptions.New()
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
