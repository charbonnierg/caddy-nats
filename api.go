// SPDX-License-Identifier: Apache-2.0

package beyond

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

// GetTLSApp can be used to load the TLS caddy app from other caddy apps.
// It makes sure the TLS app is loaded using the beyond app context before returning it.
func (b *Beyond) GetTLSApp() (*caddytls.TLS, error) {
	if b.tls != nil {
		return b.tls, nil
	}
	// Let's load the TLS app
	b.logger.Warn("loading tls app")
	unm, err := b.ctx.App("tls")
	if err != nil {
		return nil, fmt.Errorf("failed to load tls app: %v", err)
	}
	tlsApp, ok := unm.(*caddytls.TLS)
	if !ok {
		return nil, errors.New("invalid tls app module type")
	}
	b.tls = tlsApp
	return tlsApp, nil
}

// RegisterApp can be used to load the Beyond caddy app module from other apps.
// Host modules in the beyond namespace must NOT call this function.
func RegisterApp(ctx caddy.Context, app App) (*Beyond, error) {
	unm, err := ctx.App("beyond")
	if err != nil {
		return nil, fmt.Errorf("failed to load beyond module")
	}
	b, ok := unm.(*Beyond)
	if !ok {
		return nil, errors.New("invalid beyond module type")
	}
	if err := b.apps.Add(app); err != nil {
		return nil, fmt.Errorf("failed to register app: %v", err)
	}
	b.logger.Warn("registered app", zap.String("app", app.CaddyModule().ID.Name()))
	return b, nil
}

// Helper function used to load other app from other apps or from modules in the beyond namespace.
// If app is not loaded yet, it will be loaded and registered.
// If app has already been loaded, it will be returned immediately.
// It is important that apps module register themselves before calling
// LoadApp in case other apps need to use them.
func (b *Beyond) LoadApp(parent App, id string) (App, error) {
	if b.apps == nil {
		return nil, errors.New("beyond module is not loaded yet")
	}

	loaded, ok := b.apps.Get(id)
	if ok {
		b.logger.Warn("returning app from cache", zap.String("caller", parent.CaddyModule().ID.Name()), zap.String("app", id))
		return loaded, nil
	}
	b.logger.Warn("loading app", zap.String("caller", parent.CaddyModule().ID.Name()), zap.String("app", id))
	unm, err := b.ctx.App(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load app: %v", err)
	}
	_app, ok := unm.(App)
	if !ok {
		return nil, errors.New("invalid app module type")
	}
	b.logger.Warn("loaded app", zap.String("caller", parent.CaddyModule().ID.Name()), zap.String("app", id))
	return _app, nil
}
