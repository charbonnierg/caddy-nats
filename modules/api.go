// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// Reload will reload the NATS server configuration.
func (a *App) Reload() error {
	return a.runner.Reload()
}

// CreateClient will create a NATS client connected to the NATS server.
func (a *App) CreateClient(options ...nats.Option) (*nats.Conn, error) {
	srv := a.runner.Server()
	if srv == nil {
		return nil, fmt.Errorf("server is not available")
	}
	opts := []nats.Option{}
	opts = append(opts, options...)
	opts = append(opts, nats.InProcessServer(srv))
	client, err := nats.Connect("", opts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}
