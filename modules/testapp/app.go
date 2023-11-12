// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package testapp

import (
	"context"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(TestApp{})
}

func (TestApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "testapp",
		New: func() caddy.Module { return new(TestApp) },
	}
}

type TestApp struct {
	ctx    caddy.Context
	cancel context.CancelFunc
	logger *zap.Logger
	done   chan struct{}
}

// Pattern for a Provision method
func (a *TestApp) Provision(ctx caddy.Context) error {
	// Just to make sure that caddy.NewContext will preserve values added in parent context
	ctx.Context = context.WithValue(ctx.Context, "key", "value")
	// Create a new cancellable context
	a.ctx, a.cancel = caddy.NewContext(ctx)
	// Get a logger from the context
	a.logger = a.ctx.Logger()
	// Create a channel which will be closed when the app is stopped
	a.done = make(chan struct{})
	return nil
}

func (a *TestApp) Start() error {
	// Kick off a goroutine to do something
	go func() {
		for {
			select {
			// When the context is cancelled, close the done channel and return
			case <-a.ctx.Done():
				a.logger.Warn("TestApp is stopping")
				close(a.done)
				return
			// After some time passes, do something
			case <-time.After(1 * time.Second):
				value := a.ctx.Value("key")
				a.logger.Info("TestApp is running", zap.String("key", value.(string)))
			}
		}
	}()
	return nil
}

// Pattern for a stop method
func (a *TestApp) Stop() error {
	// Cancel the context
	a.cancel()
	// Wait for the done channel to be closed
	<-(a.done)
	// And that's it
	return nil
}
