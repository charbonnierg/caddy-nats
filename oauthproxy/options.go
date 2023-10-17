// SPDX-License-Identifier: Apache-2.0

package oauthproxy

import (
	"time"

	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
)

// SecretSource references an individual secret value.
// Only one source within the struct should be defined at any time.
type SecretSource struct {
	// Value expects a base64 encoded string value.
	Value []byte `json:"value,omitempty"`

	// FromEnv expects the name of an environment variable.
	FromEnv string `json:"from_env,omitempty"`

	// FromFile expects a path to a file containing the secret value.
	FromFile string `json:"from_file,omitempty"`
}

// ClaimSource allows loading a header value from a claim within the session
type ClaimSource struct {
	// Claim is the name of the claim in the session that the value should be
	// loaded from.
	Claim string `json:"claim,omitempty"`

	// Prefix is an optional prefix that will be prepended to the value of the
	// claim if it is non-empty.
	Prefix string `json:"prefix,omitempty"`

	// BasicAuthPassword converts this claim into a basic auth header.
	// Note the value of claim will become the basic auth username and the
	// basicAuthPassword will be used as the password value.
	BasicAuthPassword *SecretSource `json:"basic_auth_password,omitempty"`
}

// Headers is a list of Header objects that will be added to a request or
// response header.
type Headers []Header

// Header represents an individual header that will be added to a request or
// response header.
type Header struct {
	// Name is the header name to be used for this set of values.
	// Names should be unique within a list of Headers.
	Name string `json:"name,omitempty"`

	// PreserveRequestValue determines whether any values for this header
	// should be preserved for the request to the upstream server.
	// This option only applies to injected request headers.
	// Defaults to false (headers that match this header will be stripped).
	PreserveRequestValue bool `json:"preserve_request_value,omitempty"`

	// Values contains the desired values for this header
	Values HeaderValues `json:"values,omitempty"`
}

// HeaderValues is a list of HeaderValue objects that will be added to a
// request or response header.
type HeaderValues []HeaderValue

// HeaderValue represents a single header value and the sources that can
// make up the header value
type HeaderValue struct {
	// Allow users to load the value from a secret source
	*SecretSource `json:",omitempty"`

	// Allow users to load the value from a session claim
	*ClaimSource `json:",omitempty"`
}

// Templates includes options for configuring the sign in and error pages
// appearance.
type Templates struct {
	// Path is the path to a folder containing a sign_in.html and an error.html
	// template.
	// These files will be used instead of the default templates if present.
	// If either file is missing, the default will be used instead.
	Path string `json:"path,omitempty"`

	// CustomLogo is the path or a URL to a logo that should replace the default logo
	// on the sign_in page template.
	// Supported formats are .svg, .png, .jpg and .jpeg.
	// If URL is used the format support depends on the browser.
	// To disable the default logo, set this value to "-".
	CustomLogo string `json:"custom_logo,omitempty"`

	// Banner overides the default sign_in page banner text. If unspecified,
	// the message will give users a list of allowed email domains.
	Banner string `json:"banner,omitempty"`

	// Footer overrides the default sign_in page footer text.
	Footer string `json:"footer,omitempty"`

	// DisplayLoginForm determines whether the sign_in page should render a
	// password form if a static passwords file (htpasswd file) has been
	// configured.
	DisplayLoginForm bool `json:"display_login_form,omitempty"`

	// Debug renders detailed errors when an error page is shown.
	// It is not advised to use this in production as errors may contain sensitive
	// information.
	// Use only for diagnosing backend errors.
	Debug bool `json:"show_debug_on_error,omitempty"`
}

// Cookie contains configuration options for the cookie.
type Cookie struct {
	Name           string        `json:"name,omitempty"`
	Secret         string        `json:"secret,omitempty"`
	Domains        []string      `json:"domains,omitempty"`
	Path           string        `json:"path,omitempty"`
	Expire         time.Duration `json:"expire,omitempty"`
	Refresh        time.Duration `json:"refresh,omitempty"`
	NoSecure       bool          `json:"no_secure,omitempty"`
	NoHTTPOnly     bool          `json:"no_http_only,omitempty"`
	SameSite       string        `json:"same_site,omitempty"`
	CSRFPerRequest bool          `json:"csrf_per_request,omitempty"`
	CSRFExpire     time.Duration `json:"csrf_expire,omitempty"`
}

// Options contains all configuration options for oauth2-proxy endpoint.
type Options struct {
	ProxyPrefix             string   `json:"proxy_prefix"`
	PingPath                string   `json:"ping_path,omitempty"`
	PingUserAgent           string   `json:"ping_user_agent,omitempty"`
	ReadyPath               string   `json:"ready_path,omitempty"`
	RealClientIPHeader      string   `json:"real_client_ip_header,omitempty"`
	TrustedIPs              []string `json:"trusted_ips,omitempty"`
	RawRedirectURL          string   `json:"redirect_url,omitempty"`
	AuthenticatedEmailsFile string   `json:"authenticated_emails-file,omitempty"`
	EmailDomains            []string `json:"email_domains,omitempty"`
	WhitelistDomains        []string `json:"whitelist_domains,omitempty"`
	HtpasswdFile            string   `json:"htpasswd_file,omitempty"`
	HtpasswdUserGroups      []string `json:"htpasswd_user_groups,omitempty"`
	ReverseProxy            bool     `json:"reverse_proxy,omitempty"`

	Cookie    Cookie    `json:"cookie"`
	Templates Templates `json:"templates"`

	InjectRequestHeaders  Headers `json:"inject_request_headers,omitempty"`
	InjectResponseHeaders Headers `json:"inject_response_headers,omitempty"`

	Providers options.Providers `json:"providers"`

	APIRoutes             []string `json:"api_routes,omitempty"`
	SkipAuthRegex         []string `json:"skip_auth_regex,omitempty"`
	SkipAuthRoutes        []string `json:"skip_auth_routes,omitempty"`
	SkipJwtBearerTokens   bool     `json:"skip_jwt_bearer_tokens,omitempty"`
	ExtraJwtIssuers       []string `json:"extra_jwt_issuers,omitempty"`
	SkipProviderButton    bool     `json:"skip_provider_button,omitempty"`
	SSLInsecureSkipVerify bool     `json:"ssl_insecure_skip_verify,omitempty"`
	SkipAuthPreflight     bool     `json:"skip_auth_preflight,omitempty"`
	ForceJSONErrors       bool     `json:"force_json_errors,omitempty"`
}

func (o *Options) oauth2proxyOptions() *options.Options {
	opts := options.NewOptions()
	opts.Logging.AuthEnabled = false
	opts.Logging.StandardEnabled = false
	opts.Logging.RequestEnabled = false
	opts.ReverseProxy = o.ReverseProxy
	if o.ProxyPrefix != "" {
		opts.ProxyPrefix = o.ProxyPrefix
	}
	if o.PingPath != "" {
		opts.PingPath = o.PingPath
	}
	if o.PingUserAgent != "" {
		opts.PingUserAgent = o.PingUserAgent
	}
	if o.ReadyPath != "" {
		opts.ReadyPath = o.ReadyPath
	}
	if o.RealClientIPHeader != "" {
		opts.RealClientIPHeader = o.RealClientIPHeader
	}
	if o.TrustedIPs != nil {
		opts.TrustedIPs = o.TrustedIPs
	}
	if o.RawRedirectURL != "" {
		opts.RawRedirectURL = o.RawRedirectURL
	}
	if o.AuthenticatedEmailsFile != "" {
		opts.AuthenticatedEmailsFile = o.AuthenticatedEmailsFile
	}
	if o.EmailDomains != nil {
		opts.EmailDomains = o.EmailDomains
	}
	if o.WhitelistDomains != nil {
		opts.WhitelistDomains = o.WhitelistDomains
	}
	if o.HtpasswdFile != "" {
		opts.HtpasswdFile = o.HtpasswdFile
	}
	if o.HtpasswdUserGroups != nil {
		opts.HtpasswdUserGroups = o.HtpasswdUserGroups
	}
	if o.Cookie.Name != "" {
		opts.Cookie.Name = o.Cookie.Name
	}
	if o.Cookie.Secret != "" {
		opts.Cookie.Secret = o.Cookie.Secret
	}
	if o.Cookie.Domains != nil {
		opts.Cookie.Domains = o.Cookie.Domains
	}
	if o.Cookie.Path != "" {
		opts.Cookie.Path = o.Cookie.Path
	}
	if o.Cookie.Expire != 0 {
		opts.Cookie.Expire = o.Cookie.Expire
	}
	if o.Cookie.Refresh != 0 {
		opts.Cookie.Refresh = o.Cookie.Refresh
	}
	if o.Cookie.NoSecure {
		opts.Cookie.Secure = false
	}
	if o.Cookie.NoHTTPOnly {
		opts.Cookie.HTTPOnly = false
	}
	if o.Cookie.SameSite != "" {
		opts.Cookie.SameSite = o.Cookie.SameSite
	}
	if o.Cookie.CSRFPerRequest {
		opts.Cookie.CSRFPerRequest = o.Cookie.CSRFPerRequest
	}
	if o.Cookie.CSRFExpire != 0 {
		opts.Cookie.CSRFExpire = o.Cookie.CSRFExpire
	}
	if o.Templates.Path != "" {
		opts.Templates.Path = o.Templates.Path
	}
	if o.Templates.CustomLogo != "" {
		opts.Templates.CustomLogo = o.Templates.CustomLogo
	}
	if o.Templates.Banner != "" {
		opts.Templates.Banner = o.Templates.Banner
	}
	if o.Templates.Footer != "" {
		opts.Templates.Footer = o.Templates.Footer
	}
	if o.Templates.Debug {
		opts.Templates.Debug = o.Templates.Debug
	}
	if o.Templates.DisplayLoginForm {
		opts.Templates.DisplayLoginForm = o.Templates.DisplayLoginForm
	}
	if o.InjectRequestHeaders != nil {
		opts.InjectRequestHeaders = o.InjectRequestHeaders.oauth2proxyOptions()
	}
	if o.InjectResponseHeaders != nil {
		opts.InjectResponseHeaders = o.InjectResponseHeaders.oauth2proxyOptions()
	}
	if o.Providers != nil {
		opts.Providers = o.Providers
	}
	if o.APIRoutes != nil {
		opts.APIRoutes = o.APIRoutes
	}
	if o.SkipAuthRegex != nil {
		opts.SkipAuthRegex = o.SkipAuthRegex
	}
	if o.SkipAuthRoutes != nil {

		opts.SkipAuthRoutes = o.SkipAuthRoutes
	}
	if o.SkipJwtBearerTokens {
		opts.SkipJwtBearerTokens = o.SkipJwtBearerTokens
	}
	if o.ExtraJwtIssuers != nil {
		opts.ExtraJwtIssuers = o.ExtraJwtIssuers
	}
	if o.SkipProviderButton {
		opts.SkipProviderButton = o.SkipProviderButton
	}
	if o.SSLInsecureSkipVerify {
		opts.SSLInsecureSkipVerify = o.SSLInsecureSkipVerify
	}
	if o.SkipAuthPreflight {
		opts.SkipAuthPreflight = o.SkipAuthPreflight
	}
	if o.ForceJSONErrors {
		opts.ForceJSONErrors = o.ForceJSONErrors
	}
	return opts
}

func (c *ClaimSource) oauth2proxyOptions() options.ClaimSource {
	var basicAuthPassword options.SecretSource
	if c.BasicAuthPassword != nil {
		basicAuthPassword = options.SecretSource{
			Value:    c.BasicAuthPassword.Value,
			FromEnv:  c.BasicAuthPassword.FromEnv,
			FromFile: c.BasicAuthPassword.FromFile,
		}
	}
	return options.ClaimSource{
		Claim:             c.Claim,
		Prefix:            c.Prefix,
		BasicAuthPassword: &basicAuthPassword,
	}
}

func (h *Header) oauth2proxyOptions() *options.Header {
	return &options.Header{
		Name:                 h.Name,
		PreserveRequestValue: h.PreserveRequestValue,
		Values:               h.Values.oauth2proxyOptions(),
	}
}

func (h *HeaderValue) oauth2proxyOptions() options.HeaderValue {
	var source options.SecretSource
	if h.SecretSource != nil {
		source = options.SecretSource{
			Value:    h.Value,
			FromEnv:  h.FromEnv,
			FromFile: h.FromFile,
		}
	}
	var claimSource options.ClaimSource
	if h.ClaimSource != nil {
		claimSource = h.ClaimSource.oauth2proxyOptions()
	}
	return options.HeaderValue{
		SecretSource: &source,
		ClaimSource:  &claimSource,
	}
}

func (h Headers) oauth2proxyOptions() []options.Header {
	var headers []options.Header
	for _, header := range h {
		headers = append(headers, *header.oauth2proxyOptions())
	}
	return headers
}

func (h HeaderValues) oauth2proxyOptions() []options.HeaderValue {
	var values []options.HeaderValue
	for _, hv := range h {
		values = append(values, hv.oauth2proxyOptions())
	}
	return values
}
