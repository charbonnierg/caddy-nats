// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond"
)

// Reload will reload the NATS server configuration.
func (a *App) Reload() error {
	return a.runner.Reload()
}

func (a *App) Context() caddy.Context {
	return a.ctx
}

func (a *App) GetServer() (*server.Server, error) {
	if a.runner == nil {
		return nil, fmt.Errorf("server is not available")
	}
	srv := a.runner.Server()
	if srv == nil {
		return nil, fmt.Errorf("server is not available")
	}
	if !srv.ReadyForConnections(5) {
		return nil, fmt.Errorf("server is not ready for connections")
	}
	return srv, nil
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

func (a *App) LoadApp(id string) (beyond.App, error) {
	if a.beyond == nil {
		return nil, fmt.Errorf("beyond is not available")
	}
	return a.beyond.LoadApp(a, id)
}
