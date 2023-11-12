// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package endpoint

import (
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/quara-dev/beyond/modules/oauth2"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (ep *Endpoint) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "store":
			if !d.NextArg() {
				return d.Err("expected store type")
			}
			storeTypeShort := d.Val()
			storeType := "oauth2.stores." + storeTypeShort
			store, err := caddyfile.UnmarshalModule(d, storeType)
			if err != nil {
				return err
			}
			s, ok := store.(oauth2.Store)
			if !ok {
				return d.Errf("invalid session store type: %T", store)
			}
			ep.Store = caddyconfig.JSONModuleObject(s, "type", storeTypeShort, nil)
		case "proxy_prefix":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.ProxyPrefix) {
				return d.ArgErr()
			}
		case "ping_path":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.PingPath) {
				return d.ArgErr()
			}
		case "ping_user_agent":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.PingUserAgent) {
				return d.ArgErr()
			}
		case "ready_path":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.ReadyPath) {
				return d.ArgErr()
			}
		case "real_client_ip_header":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.RealClientIPHeader) {
				return d.ArgErr()
			}
		case "trusted_ips":
			makeOptions(ep)
			if ep.Options.TrustedIPs == nil {
				ep.Options.TrustedIPs = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.TrustedIPs = append(ep.Options.TrustedIPs, val)
				}
			}
		case "redirect_url":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.RawRedirectURL) {
				return d.ArgErr()
			}
		case "authenticated_emails_file":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.AuthenticatedEmailsFile) {
				return d.ArgErr()
			}
		case "email_domains":
			makeOptions(ep)
			if ep.Options.EmailDomains == nil {
				ep.Options.EmailDomains = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.EmailDomains = append(ep.Options.EmailDomains, val)
				}
			}
		case "whitelist_domains":
			makeOptions(ep)
			if ep.Options.WhitelistDomains == nil {
				ep.Options.WhitelistDomains = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.WhitelistDomains = append(ep.Options.WhitelistDomains, val)
				}
			}
		case "htpasswd_file":
			makeOptions(ep)
			if !d.AllArgs(&ep.Options.HtpasswdFile) {
				return d.ArgErr()
			}
		case "htpasswd_user_groups":
			makeOptions(ep)
			if ep.Options.HtpasswdUserGroups == nil {
				ep.Options.HtpasswdUserGroups = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.HtpasswdUserGroups = append(ep.Options.HtpasswdUserGroups, val)
				}
			}
		case "reverse_proxy":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.ReverseProxy)
			if err != nil {
				return err
			}
		case "cookie_name":
			makeOptionsCookie(ep)
			if !d.AllArgs(&ep.Options.Cookie.Name) {
				return d.ArgErr()
			}
		case "cookie_secret":
			makeOptionsCookie(ep)
			if !d.AllArgs(&ep.Options.Cookie.Secret) {
				return d.ArgErr()
			}
		case "cookie_domains":
			makeOptionsCookie(ep)
			if ep.Options.Cookie.Domains == nil {
				ep.Options.Cookie.Domains = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.Cookie.Domains = append(ep.Options.Cookie.Domains, val)
				}
			}
		case "cookie_expire":
			makeOptionsCookie(ep)
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid duration: %v", err)
			}
			ep.Options.Cookie.Expire = val
		case "cookie_refresh":
			makeOptionsCookie(ep)
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid duration: %v", err)
			}
			ep.Options.Cookie.Refresh = val
		case "cookie_secure":
			makeOptionsCookie(ep)
			err := parser.ParseBool(d, &ep.Options.Cookie.NoSecure, parser.Reverse())
			if err != nil {
				return err
			}
		case "cookie_http_only":
			makeOptionsCookie(ep)
			err := parser.ParseBool(d, &ep.Options.Cookie.NoHTTPOnly, parser.Reverse())
			if err != nil {
				return err
			}
		case "cookie_same_site":
			makeOptionsCookie(ep)
			if !d.AllArgs(&ep.Options.Cookie.SameSite) {
				return d.ArgErr()
			}
		case "csrf_per_request":
			makeOptionsCookie(ep)
			err := parser.ParseBool(d, &ep.Options.Cookie.CSRFPerRequest)
			if err != nil {
				return err
			}
		case "csrf_expire":
			makeOptionsCookie(ep)
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid duration: %v", err)
			}
			ep.Options.Cookie.CSRFExpire = val
		case "templates_path":
			makeOptionsTemplates(ep)
			if !d.AllArgs(&ep.Options.Templates.Path) {
				return d.ArgErr()
			}
		case "custom_logo":
			makeOptionsTemplates(ep)
			if !d.AllArgs(&ep.Options.Templates.CustomLogo) {
				return d.ArgErr()
			}
		case "templates_banner":
			makeOptionsTemplates(ep)
			if !d.AllArgs(&ep.Options.Templates.Banner) {
				return d.ArgErr()
			}
		case "templates_footer":
			makeOptionsTemplates(ep)
			if !d.AllArgs(&ep.Options.Templates.Footer) {
				return d.ArgErr()
			}
		case "display_login_form":
			makeOptionsTemplates(ep)
			err := parser.ParseBool(d, &ep.Options.Templates.DisplayLoginForm)
			if err != nil {
				return err
			}
		case "display_debug":
			makeOptionsTemplates(ep)
			err := parser.ParseBool(d, &ep.Options.Templates.Debug)
			if err != nil {
				return err
			}
		case "inject_request_header":
			makeOptions(ep)
			if ep.Options.InjectRequestHeaders == nil {
				ep.Options.InjectRequestHeaders = Headers{}
			}
			return d.Err("not implemented")
		case "inject_response_header":
			makeOptions(ep)
			if ep.Options.InjectResponseHeaders == nil {
				ep.Options.InjectResponseHeaders = Headers{}
			}
			return d.Err("not implemented")
		case "provider":
			makeOptions(ep)
			if ep.Options.Providers == nil {
				ep.Options.Providers = []options.Provider{}
			}
			p := options.Provider{}
			if !d.NextArg() {
				return d.Err("expected provider type")
			}
			p.Type = options.ProviderType(d.Val())
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "client_id":
					if !d.AllArgs(&p.ClientID) {
						return d.ArgErr()
					}
				case "client_secret":
					if !d.AllArgs(&p.ClientSecret) {
						return d.ArgErr()
					}
				case "client_secret_file":
					if !d.AllArgs(&p.ClientSecretFile) {
						return d.ArgErr()
					}
				case "name":
					if !d.AllArgs(&p.Name) {
						return d.ArgErr()
					}
				case "id":
					if !d.AllArgs(&p.ID) {
						return d.ArgErr()
					}
				case "ca_files":
					if p.CAFiles == nil {
						p.CAFiles = []string{}
					}
					for d.NextArg() {
						if val := d.Val(); val != "" {
							p.CAFiles = append(p.CAFiles, val)
						}
					}
				case "login_url":
					if !d.AllArgs(&p.LoginURL) {
						return d.ArgErr()
					}
				case "redeem_url":
					if !d.AllArgs(&p.RedeemURL) {
						return d.ArgErr()
					}
				case "profile_url":
					if !d.AllArgs(&p.ProfileURL) {
						return d.ArgErr()
					}
				case "protected_resource":
					if !d.AllArgs(&p.ProtectedResource) {
						return d.ArgErr()
					}
				case "validate_url":
					if !d.AllArgs(&p.ValidateURL) {
						return d.ArgErr()
					}
				case "scope":
					scopes := []string{}
					for d.NextArg() {
						if val := d.Val(); val != "" {
							scopes = append(scopes, val)
						}
					}
					p.Scope = strings.Join(scopes, " ")
				case "allowed_groups":
					if p.AllowedGroups == nil {
						p.AllowedGroups = []string{}
					}
					for d.NextArg() {
						if val := d.Val(); val != "" {
							p.AllowedGroups = append(p.AllowedGroups, val)
						}
					}
				case "code_challenge_method":
					if !d.AllArgs(&p.CodeChallengeMethod) {
						return d.ArgErr()
					}
				case "oidc_issuer_url":
					if !d.AllArgs(&p.OIDCConfig.IssuerURL) {
						return d.ArgErr()
					}
				case "oidc_jwks_url":
					if !d.AllArgs(&p.OIDCConfig.JwksURL) {
						return d.ArgErr()
					}
				case "oidc_insecure_allow_unverified_email":
					err := parser.ParseBool(d, &p.OIDCConfig.InsecureAllowUnverifiedEmail)
					if err != nil {
						return err
					}
				case "oidc_insecure_skip_issuer_verification":
					err := parser.ParseBool(d, &p.OIDCConfig.InsecureSkipIssuerVerification)
					if err != nil {
						return err
					}
				case "oidc_insecure_skip_nonce":
					err := parser.ParseBool(d, &p.OIDCConfig.InsecureSkipNonce)
					if err != nil {
						return err
					}
				case "oidc_skip_discovery":
					err := parser.ParseBool(d, &p.OIDCConfig.SkipDiscovery)
					if err != nil {
						return err
					}
				case "oidc_email_claim":
					if !d.AllArgs(&p.OIDCConfig.EmailClaim) {
						return d.ArgErr()
					}
				case "oidc_groups_claim":
					if !d.AllArgs(&p.OIDCConfig.GroupsClaim) {
						return d.ArgErr()
					}
				case "oidc_user_id_claim":
					if !d.AllArgs(&p.OIDCConfig.UserIDClaim) {
						return d.ArgErr()
					}
				case "oidc_audience_claims":
					if p.OIDCConfig.AudienceClaims == nil {
						p.OIDCConfig.AudienceClaims = []string{}
					}
					for d.NextArg() {
						if val := d.Val(); val != "" {
							p.OIDCConfig.AudienceClaims = append(p.OIDCConfig.AudienceClaims, val)
						}
					}
				case "oidc_extra_audiences":
					if p.OIDCConfig.ExtraAudiences == nil {
						p.OIDCConfig.ExtraAudiences = []string{}
					}
					for d.NextArg() {
						if val := d.Val(); val != "" {
							p.OIDCConfig.ExtraAudiences = append(p.OIDCConfig.ExtraAudiences, val)
						}
					}
				default:
					return d.Errf("unrecognized subdirective: %s", d.Val())
				}
			}
			ep.Options.Providers = append(ep.Options.Providers, p)
		case "api_routes":
			makeOptions(ep)
			if ep.Options.APIRoutes == nil {
				ep.Options.APIRoutes = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.APIRoutes = append(ep.Options.APIRoutes, val)

				}
			}
		case "skip_auth_regex":
			makeOptions(ep)
			if ep.Options.SkipAuthRegex == nil {
				ep.Options.SkipAuthRegex = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.SkipAuthRegex = append(ep.Options.SkipAuthRegex, val)
				}
			}
		case "skip_auth_routes":
			makeOptions(ep)
			if ep.Options.SkipAuthRoutes == nil {
				ep.Options.SkipAuthRoutes = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.SkipAuthRoutes = append(ep.Options.SkipAuthRoutes, val)
				}
			}
		case "skip_jwt_bearer_tokens":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.SkipJwtBearerTokens)
			if err != nil {
				return err
			}
		case "extra_jwt_issuers":
			makeOptions(ep)
			if ep.Options.ExtraJwtIssuers == nil {
				ep.Options.ExtraJwtIssuers = []string{}
			}
			for d.NextArg() {
				if val := d.Val(); val != "" {
					ep.Options.ExtraJwtIssuers = append(ep.Options.ExtraJwtIssuers, val)
				}
			}
		case "skip_provider_button":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.SkipProviderButton)
			if err != nil {
				return err
			}
		case "ssl_insecure_skip_verify":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.SSLInsecureSkipVerify)
			if err != nil {
				return err
			}
		case "skip_auth_preflight":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.SkipAuthPreflight)
			if err != nil {
				return err
			}
		case "force_json_errors":
			makeOptions(ep)
			err := parser.ParseBool(d, &ep.Options.ForceJSONErrors)
			if err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}

func makeOptions(ep *Endpoint) {
	if ep.Options == nil {
		ep.Options = &Options{}
	}
}

func makeOptionsTemplates(ep *Endpoint) {
	makeOptions(ep)
	if ep.Options.Templates == nil {
		ep.Options.Templates = &Templates{}
	}
}

func makeOptionsCookie(ep *Endpoint) {
	makeOptions(ep)
	if ep.Options.Cookie == nil {
		ep.Options.Cookie = &Cookie{}
	}
}
