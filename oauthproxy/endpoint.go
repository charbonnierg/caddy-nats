package oauthproxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/encryption"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/validation"
	"github.com/oauth2-proxy/oauth2-proxy/v7/server"
	"go.uber.org/zap"
)

type Endpoint struct {
	logger           *zap.Logger
	cipher           encryption.Cipher
	opts             *options.Options
	proxy            *server.OAuthProxy
	Name             string             `json:"name,omitempty"`
	Providers        []options.Provider `json:"providers,omitempty"`
	ExtraJwtIssuers  []string           `json:"extra_jwt_issuers,omitempty"`
	CookieDomains    []string           `json:"cookie_domains,omitempty"`
	WhitelistDomains []string           `json:"whitelist_domains,omitempty"`
}

func (Endpoint) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauthproxy.endpoint",
		New: func() caddy.Module { return new(Endpoint) },
	}
}

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	r = r.WithContext(context.WithValue(r.Context(), nextKey{}, next))
	e.proxy.ServeHTTP(w, r)
	return nil
}

func (e *Endpoint) Provision(ctx caddy.Context) error {
	// Initialize options
	if e.opts == nil {
		e.opts = options.NewOptions()
	}

	// Save logger
	e.logger = ctx.Logger().Named(e.Name)

	// TODO: The secret is randomely generated, but it is not persisted
	if e.opts.Cookie.Secret == "" {
		secret, err := generateRandomASCIIString(32)
		if err != nil {
			return err
		}
		e.opts.Cookie.Secret = secret
	}
	// Configure cookie name
	if e.opts.Cookie.Name == "" {
		e.opts.Cookie.Name = fmt.Sprintf("%s_%s", e.Name, "oauth2_proxy")
	}
	// Configure providers
	e.opts.Providers = []options.Provider{}
	e.opts.Providers = append(e.opts.Providers, e.Providers...)

	// Configure remaining options
	e.opts.ExtraJwtIssuers = append(e.opts.ExtraJwtIssuers, e.ExtraJwtIssuers...)
	e.opts.Cookie.Domains = e.CookieDomains
	e.opts.WhitelistDomains = e.WhitelistDomains
	e.opts.SkipJwtBearerTokens = true
	e.opts.ReverseProxy = true
	e.opts.SkipProviderButton = true
	e.opts.EmailDomains = []string{"*"}

	// This is a hack to get the access token and id token to the backend
	e.opts.InjectRequestHeaders = append(e.opts.InjectRequestHeaders, options.Header{
		Name: "X-Forwarded-Access-Token",
		Values: []options.HeaderValue{
			{
				ClaimSource: &options.ClaimSource{
					Claim: "access_token",
				},
			},
		},
	},
		options.Header{
			Name: "Authorization",
			Values: []options.HeaderValue{
				{
					ClaimSource: &options.ClaimSource{
						Claim: "id_token",
					},
				},
			},
		})

	// Save cipher
	cipher, err := encryption.NewCFBCipher(encryption.SecretBytes(e.opts.Cookie.Secret))
	if err != nil {
		return fmt.Errorf("error initialising cipher: %v", err)
	}
	e.cipher = cipher

	// Validate options
	if err := validation.Validate(e.opts); err != nil {
		return err
	}

	return nil
}

func (e *Endpoint) setup() error {
	chainer := chainer{logger: e.logger}
	validator := server.NewValidator(e.opts.EmailDomains, e.opts.AuthenticatedEmailsFile)
	proxy, err := server.NewEmbeddedOauthProxy(e.opts, validator, &chainer)
	if err != nil {
		return err
	}
	e.proxy = proxy
	return nil
}

type nextKey struct{}

type chainer struct {
	logger *zap.Logger
}

func (h chainer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("serving authorized request", zap.Any("access_token", r.Header["X-Forwarded-Access-Token"]), zap.Any("id_token", r.Header["Authorization"]))
	nextRaw := r.Context().Value(nextKey{})
	if nextRaw == nil {
		h.logger.Error("next handler not found in request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	next, ok := nextRaw.(caddyhttp.Handler)
	if !ok {
		h.logger.Error("next handler is not an http.Handler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err := next.ServeHTTP(w, r)
	if err != nil {
		h.logger.Error("error serving next handler", zap.Error(err))
	}
}
