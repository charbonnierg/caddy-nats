// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/oauth2"
	"github.com/quara-dev/beyond/modules/oauth2/endpoint"
	"go.uber.org/zap"
)

// Context returns the caddy context for the app.
func (a *App) Context() caddy.Context { return a.ctx }

// Logger returns the logger for the app.
func (a *App) Logger() *zap.Logger { return a.logger }

// GetReplacer returns the caddy replacer for the app.
func (a *App) GetReplacer() *caddy.Replacer { return a.repl }

// LoadBeyondApp returns the beyond app with the given id.
func (a *App) LoadBeyondApp(id string) (beyond.App, error) {
	if a.beyond == nil {
		return nil, fmt.Errorf("beyond is not available")
	}
	return a.beyond.LoadApp(id)
}

// GetLogger returns a child logger with the given name.
// The logger is created using the app logger.
func (a *App) GetLogger(name string) *zap.Logger {
	return a.logger.Named(name)
}

// GetOrAddEndpoint returns the endpoint with the given name, or adds it to the app if it does not exist.
func (a *App) getOrAddEndpoint(e oauth2.Endpoint) (oauth2.Endpoint, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, existing := range a.endpoints {
		if e.Name() == existing.Name() {
			if e.IsReference() {
				return existing, nil
			}
			if !existing.Equals(e) {
				return nil, fmt.Errorf("endpoint %s already exists with different configuration", e.Name())
			}
			return existing, nil
		}
	}
	return e, a.addEndpoint(e)
}

// GetOrAddEndpoint returns the endpoint with the given name, or adds it to the app if it does not exist.
// If the endpoint already exists with different configuration, an error is returned.
// If the endpoint is added to the app, it is provisioned before being returned.
// An error is returned when the endpoint does not exist yet and cannot be provisioned.
func (a *App) GetOrAddEndpoint(e oauth2.Endpoint) (oauth2.Endpoint, error) {
	ep, ok := e.(*endpoint.Endpoint)
	if !ok {
		return nil, fmt.Errorf("invalid endpoint type")
	}
	return a.getOrAddEndpoint(ep)
}

// GetEndpoint returns the endpoint with the given name, or an error if it does not exist.
// Error message is prefixed with "unknown oauth2 endpoint: ".
// Endpoints must be provisioned using AddEnpoint method before they can be retrieved.
func (a *App) GetEndpoint(name string) (oauth2.Endpoint, error) {
	return a.getEndpoint(name)
}

// getEndpoint returns the endpoint with the given name, or an error if it does not exist.
func (a *App) getEndpoint(name string) (oauth2.Endpoint, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, e := range a.endpoints {
		if e.Name() == name {
			return e, nil
		}
	}
	return nil, fmt.Errorf("unknown oauth2 endpoint: %s", name)
}

// addEndpoint adds the given endpoint to the app.
// If an endpoint with the same name already exists, an error is returned.
// Otherwise, the endpoint is provisioned and added to the app.
// It can later be retrieved using GetEndpoint method.
func (a *App) addEndpoint(e oauth2.Endpoint) error {
	for _, endpoint := range a.endpoints {
		if endpoint.Name() == e.Name() {
			return fmt.Errorf("endpoint %s already exists", e.Name())
		}
	}
	if err := e.Provision(a); err != nil {
		return err
	}
	a.endpoints = append(a.endpoints, e)
	return nil
}
