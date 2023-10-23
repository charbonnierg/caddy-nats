// SPDX-License-Identifier: Apache-2.0

package beyond

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"go.uber.org/zap"
)

// Register Beyond as a Caddy module on import.
func init() {
	caddy.RegisterModule(new(Beyond))
}

// Beyond is the top-level module for all beyond caddy modules.
// It implements the caddy.App interface.
type Beyond struct {
	logger *zap.Logger
	ctx    *caddy.Context
	tls    *caddytls.TLS
	apps   Apps
	// Problem with this approach: we cannot ship a binary without NATS
	// This DOES NOT WORK !!
	// Nats    *nats.App    `json:"nats,omitempty"`
	// OAuth2  *oauth2.App  `json:"oauth2,omitempty"`
	// Secrets *secrets.App `json:"secrets,omitempty"`
}

// CaddyModule returns the Caddy module information.
// It is required to implement the caddy.Module interface.
func (Beyond) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "beyond",
		New: func() caddy.Module { return new(Beyond) },
	}
}

// Start will start the Beyond module. It is required to implement the caddy.App interface.
func (b *Beyond) Start() error { return nil }

// Stop will stop the Beyond module. It is required to implement the caddy.App interface.
func (b *Beyond) Stop() error { return nil }

// Provision will configure and validate the Beyond module.
// It is required to implement the caddy.Provisioner interface.
func (b *Beyond) Provision(ctx caddy.Context) error {
	// Let's save the context for later
	b.ctx = &ctx
	// Let's get the root logger for all Beyond modules from the context
	b.logger = ctx.Logger()
	// Initialize apps
	b.apps = make(Apps)
	return nil
}

// Interface guards
var (
	_ caddy.App         = (*Beyond)(nil)
	_ caddy.Provisioner = (*Beyond)(nil)
)
