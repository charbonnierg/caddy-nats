package http_handler

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
)

// ParsePublishHandler parses the nats_publish directive. Syntax:
//
//	oauth2_proxy {
//
// }
func ParseOauth2ProxyDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var p = Proxy{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

func (p *Proxy) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&p.EndpointRaw.Name) {
			return d.ArgErr()
		}
		ep := &p.EndpointRaw
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "cookie_domains":
				remainings := d.RemainingArgs()
				ep.CookieDomains = []string{}
				for _, remaining := range remainings {
					if remaining != "" {
						ep.CookieDomains = append(ep.CookieDomains, remaining)
					}
				}
			case "whitelist_domains":
				reremainings := d.RemainingArgs()
				ep.WhitelistDomains = []string{}
				for _, reremaining := range reremainings {
					if reremaining != "" {
						ep.WhitelistDomains = append(ep.WhitelistDomains, reremaining)
					}
				}
			case "provider":
				if !d.NextArg() {
					return d.ArgErr()
				}
				var providerType options.ProviderType
				switch d.Val() {
				case "azure":
					providerType = options.AzureProvider
				case "github":
					providerType = options.GitHubProvider
				case "gitlab":
					providerType = options.GitLabProvider
				case "google":
					providerType = options.GoogleProvider
				case "keycloack":
					providerType = options.KeycloakProvider
				case "oidc":
					providerType = options.OIDCProvider
				default:
					return d.Errf("unrecognized provider type %s", d.Val())
				}
				provider := options.Provider{
					Type: providerType,
				}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "id":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.ID = d.Val()
						if provider.Name == "" {
							provider.Name = provider.ID
						}
					case "name":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.Name = d.Val()
						if provider.ID == "" {
							provider.ID = provider.Name
						}
					case "client_id":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.ClientID = d.Val()
					case "client_secret":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.ClientSecret = d.Val()
					case "oidc_issuer_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.OIDCConfig.IssuerURL = d.Val()
					case "oidc_jwks_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.OIDCConfig.JwksURL = d.Val()
					case "azure_tenant":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.AzureConfig.Tenant = d.Val()
					case "scope":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.Scope = d.Val()

					case "validate_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.ValidateURL = d.Val()
					case "profile_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.ProfileURL = d.Val()
					case "login_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.LoginURL = d.Val()
					case "redeem_url":
						if !d.NextArg() {
							return d.ArgErr()
						}
						provider.RedeemURL = d.Val()
					default:
						return d.Errf("unrecognized subdirective %s", d.Val())
					}
				}

			}
		}
	}
	return nil
}