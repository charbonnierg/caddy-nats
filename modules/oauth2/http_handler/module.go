// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package http_handler

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/quara-dev/beyond/modules/oauth2"
	"github.com/quara-dev/beyond/modules/oauth2/oauth2app"
)

func init() {
	caddy.RegisterModule(OAuth2Session{})
	httpcaddyfile.RegisterHandlerDirective("oauth2_session", ParseOauth2ProxyDirective)
}

// OAuth2Session is a Caddy module that represents an oauth2 middleware endpoint.
// It implements the caddyhttp.MiddlewareHandler interface.
type OAuth2Session struct {
	endpoint    oauth2.OAuth2Endpoint
	EndpointRaw *oauth2app.Endpoint `json:"endpoint,omitempty"`
}

// CaddyModule returns the Caddy module information.
// It implements the caddy.Module interface.
func (OAuth2Session) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.oauth2_session",
		New: func() caddy.Module { return new(OAuth2Session) },
	}
}

// Provision configures the module.
// It implements the caddy.Provisioner interface.
// It is called when the module is provisioned on first load or on config change.
// It will get or add the oauth2 endpoint to the app.
func (p *OAuth2Session) Provision(ctx caddy.Context) error {
	unm, err := ctx.App("oauth2")
	if err != nil {
		return fmt.Errorf("failed to load oauth2 app module: %v", err)
	}
	app, ok := unm.(*oauth2app.App)
	if !ok {
		return fmt.Errorf("invalid oauth2 app module")
	}
	if p.EndpointRaw.Name == "" {
		return fmt.Errorf("missing endpoint name")
	}
	ep, err := app.GetOrAddEndpoint(p.EndpointRaw)
	if err != nil {
		return err
	}
	p.endpoint = ep
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
// It simply delegates the request to the endpoint handler.
func (p OAuth2Session) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return p.endpoint.ServeHTTP(w, r, next)
}

var (
	_ caddy.Provisioner           = (*OAuth2Session)(nil)
	_ caddyfile.Unmarshaler       = (*OAuth2Session)(nil)
	_ caddyhttp.MiddlewareHandler = (*OAuth2Session)(nil)
)
