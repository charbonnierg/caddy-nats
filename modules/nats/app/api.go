// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/nats/auth/policies"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
	"go.uber.org/zap"
)

// Context is the context for the NATS app.
func (a *App) Context() caddy.Context { return a.ctx }

// Logger will return the logger for the NATS app.
func (a *App) Logger() *zap.Logger { return a.logger }

// GetOptions will return the NATS server options for the NATS app.
func (a *App) GetOptions() *embedded.Options { return a.ServerOptions }

// GetServer will return the NATS server embedded in the NATS app.
func (a *App) GetServer() (*server.Server, error) {
	if a.runner == nil {
		return nil, fmt.Errorf("server is not available")
	}
	srv := a.runner.Server()
	if srv == nil {
		return nil, fmt.Errorf("server is not available")
	}
	timeout := a.ReadyTimeout
	if timeout == 0 {
		timeout = time.Second * 5
	}
	if !srv.ReadyForConnections(timeout) {
		return nil, fmt.Errorf("server is not ready for connections")
	}
	return srv, nil
}

// Reload will reload the embedded NATS server configuration.
func (a *App) ReloadServer() error { return a.runner.Reload() }

// LoadBeyondApp will load a Beyond app by its ID.
func (a *App) LoadBeyondApp(id string) (beyond.App, error) {
	if a.beyond == nil {
		return nil, fmt.Errorf("beyond is not available")
	}
	return a.beyond.LoadApp(id)
}

// AddNewTokenBasedAuthPolicy will add a new token based auth policy to the NATS server.
func (a *App) AddNewTokenBasedAuthPolicy(account string) (string, error) {
	// Need to provision on-the-fly credentials for this account
	if a.AuthService == nil {
		return "", fmt.Errorf("cannot provision token based auth policy for account '%s' without auth service", account)
	}
	token := account + "-token"
	a.AuthService.Policies = append(a.AuthService.Policies, &policies.ConnectionPolicy{
		HandlerRaw: json.RawMessage(`{"module": "allow", "account": "` + account + `"}`),
		MatchersRaw: []json.RawMessage{
			json.RawMessage(`{"type": "connect_opts", "token": "` + token + `"}`),
		},
	})
	return token, nil
}
