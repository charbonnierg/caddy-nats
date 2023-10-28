// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package beyond

import (
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

// Register can be used to load the Beyond caddy app module and register
// a new beyond app. It returns the loaded Beyond module and an error if any.
func Register(ctx caddy.Context, app App) (*Beyond, error) {
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

// LoadTLSApp can be used to load the TLS caddy app from other caddy apps.
// It makes sure the TLS app is loaded using the beyond app context before returning it.
func (b *Beyond) LoadTLSApp() (*caddytls.TLS, error) {
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

// LoadApp can be used to load other beyond apps.
// If app is not loaded yet, it will be loaded and registered.
// If app has already been loaded, it will be returned immediately.
// It is important that apps module register themselves before calling
// LoadApp in case other apps need to use them.
func (b *Beyond) LoadApp(id string) (App, error) {
	if b.apps == nil {
		return nil, errors.New("beyond module is not loaded yet")
	}
	loaded, ok := b.apps.Get(id)
	if ok {
		return loaded, nil
	}
	unm, err := b.ctx.App(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load app: %v", err)
	}
	_app, ok := unm.(App)
	if !ok {
		return nil, errors.New("invalid app module type")
	}
	return _app, nil
}

// LoadCaddyApp can be used to load any caddy app from other caddy apps.
func (b *Beyond) LoadCaddyApp(id string) (caddy.App, error) {
	if b.apps == nil {
		return nil, errors.New("beyond module is not loaded yet")
	}
	unm, err := b.ctx.App(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load app: %v", err)
	}
	app, ok := unm.(caddy.App)
	if !ok {
		return nil, errors.New("invalid app module type")
	}
	return app, nil
}
