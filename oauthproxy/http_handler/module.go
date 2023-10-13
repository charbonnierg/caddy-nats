package http_handler

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/charbonnierg/caddy-nats/oauthproxy"
)

func init() {
	caddy.RegisterModule(Proxy{})
	httpcaddyfile.RegisterHandlerDirective("oauth2_proxy", ParseOauth2ProxyDirective)
}

type Proxy struct {
	endpoint    *oauthproxy.Endpoint
	EndpointRaw oauthproxy.Endpoint `json:"endpoint,omitempty"`
}

func (Proxy) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.oauth2_proxy",
		New: func() caddy.Module { return new(Proxy) },
	}
}

func (p *Proxy) Provision(ctx caddy.Context) error {
	app, err := oauthproxy.LoadApp(ctx)
	if err != nil {
		return err
	}
	if p.EndpointRaw.Name == "" {
		return fmt.Errorf("missing endpoint name")
	}
	endpoint := app.GetEndpoint(p.EndpointRaw.Name)
	if endpoint == nil {
		err := app.AddEndpoint(&p.EndpointRaw)
		if err != nil {
			return err
		}
		p.endpoint = &p.EndpointRaw
	} else {
		p.endpoint = endpoint
	}
	return nil
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return p.endpoint.ServeHTTP(w, r, next)
}

var (
	_ caddy.Provisioner           = (*Proxy)(nil)
	_ caddyfile.Unmarshaler       = (*Proxy)(nil)
	_ caddyhttp.MiddlewareHandler = (*Proxy)(nil)
)
