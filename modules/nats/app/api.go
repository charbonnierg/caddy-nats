// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/nats/embedded"
	"go.uber.org/zap"
)

// Context is the context for the NATS app.
func (a *App) Context() caddy.Context { return a.ctx }

// Logger will return the logger for the NATS app.
func (a *App) Logger() *zap.Logger { return a.logger }

// Options will return the NATS server options for the NATS app.
func (a *App) Options() *embedded.Options { return a.options }

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

// GetAuthUserPass will return the user and password to use for the auth service
// according to server configuration.
func (a *App) GetAuthUserPass() (string, string, error) {
	// The goal is to "guess" the user and password to use for the auth callout
	if a.options.Authorization != nil {
		auth := a.options.Authorization
		accs := a.options.Accounts
		config := auth.AuthCallout
		if config != nil && config.AuthUsers != nil {
			if auth.Users != nil {
				for _, user := range auth.Users {
					for _, authUser := range config.AuthUsers {
						if user.User == authUser {
							return user.User, user.Password, nil
						}
					}
				}
			} else {
				for _, acc := range accs {
					if acc.Name == config.Account {
						for _, user := range acc.Users {
							for _, authUser := range config.AuthUsers {
								if user.User == authUser {
									return user.User, user.Password, nil
								}
							}
						}
					}
				}
			}
		}
	}
	return "", "", fmt.Errorf("user not found")
}
