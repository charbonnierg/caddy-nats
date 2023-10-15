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

// CookieStoreOptions contains configuration options for the CookieSessionStore.
type CookieStoreOptions struct {
	Minimal bool `json:"session_cookie_minimal"`
}

// RedisStoreOptions contains configuration options for the RedisSessionStore.
type RedisStoreOptions struct {
	ConnectionURL          string   `json:"connection_url"`
	Password               string   `json:"password"`
	UseSentinel            bool     `json:"use_sentinel"`
	SentinelPassword       string   `json:"sentinel_password"`
	SentinelMasterName     string   `json:"sentinel_master_name"`
	SentinelConnectionURLs []string `json:"sentinel_connection_urls"`
	UseCluster             bool     `json:"use_cluster"`
	ClusterConnectionURLs  []string `json:"cluster_connection_urls"`
	CAPath                 string   `json:"ca_path"`
	InsecureSkipTLSVerify  bool     `json:"insecure_skip_tls_verify"`
	IdleTimeout            int      `json:"idle_timeout"`
}

// SessionOptions contains configuration options for the session store.
type SessionOptions struct {
	Type   string             `json:"type"`
	Cookie CookieStoreOptions `json:"cookie"`
	Redis  RedisStoreOptions  `json:"redis"`
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

	Cookie    Cookie         `json:"cookie"`
	Session   SessionOptions `json:"session"`
	Templates Templates      `json:"templates"`

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

func (o *Options) getOptions() *options.Options {
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
	if o.Session.Type != "" {
		opts.Session.Type = o.Session.Type
	}
	if o.Session.Cookie.Minimal {
		opts.Session.Cookie.Minimal = o.Session.Cookie.Minimal
	}
	if o.Session.Redis.ConnectionURL != "" {
		opts.Session.Redis.ConnectionURL = o.Session.Redis.ConnectionURL
	}
	if o.Session.Redis.Password != "" {
		opts.Session.Redis.Password = o.Session.Redis.Password
	}
	if o.Session.Redis.UseSentinel {
		opts.Session.Redis.UseSentinel = o.Session.Redis.UseSentinel
	}
	if o.Session.Redis.SentinelPassword != "" {
		opts.Session.Redis.SentinelPassword = o.Session.Redis.SentinelPassword
	}
	if o.Session.Redis.SentinelMasterName != "" {
		opts.Session.Redis.SentinelMasterName = o.Session.Redis.SentinelMasterName
	}
	if o.Session.Redis.SentinelConnectionURLs != nil {
		opts.Session.Redis.SentinelConnectionURLs = o.Session.Redis.SentinelConnectionURLs
	}
	if o.Session.Redis.UseCluster {
		opts.Session.Redis.UseCluster = o.Session.Redis.UseCluster
	}
	if o.Session.Redis.ClusterConnectionURLs != nil {
		opts.Session.Redis.ClusterConnectionURLs = o.Session.Redis.ClusterConnectionURLs
	}
	if o.Session.Redis.CAPath != "" {
		opts.Session.Redis.CAPath = o.Session.Redis.CAPath
	}
	if o.Session.Redis.InsecureSkipTLSVerify {
		opts.Session.Redis.InsecureSkipTLSVerify = o.Session.Redis.InsecureSkipTLSVerify
	}
	if o.Session.Redis.IdleTimeout != 0 {
		opts.Session.Redis.IdleTimeout = o.Session.Redis.IdleTimeout
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
		opts.InjectRequestHeaders = o.InjectRequestHeaders.oauth2proxyHeaders()
	}
	if o.InjectResponseHeaders != nil {
		opts.InjectResponseHeaders = o.InjectResponseHeaders.oauth2proxyHeaders()
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

func (o *Options) equals(other *Options) bool {
	if len(o.ExtraJwtIssuers) != len(other.ExtraJwtIssuers) {
		return false
	}
	for i, issuer := range o.ExtraJwtIssuers {
		if issuer != other.ExtraJwtIssuers[i] {
			return false
		}
	}
	if len(o.SkipAuthRegex) != len(other.SkipAuthRegex) {
		return false
	}
	for i, regex := range o.SkipAuthRegex {
		if regex != other.SkipAuthRegex[i] {
			return false
		}
	}
	if len(o.SkipAuthRoutes) != len(other.SkipAuthRoutes) {
		return false
	}
	for i, route := range o.SkipAuthRoutes {
		if route != other.SkipAuthRoutes[i] {
			return false

		}
	}
	if len(o.APIRoutes) != len(other.APIRoutes) {
		return false
	}
	for i, route := range o.APIRoutes {
		if route != other.APIRoutes[i] {
			return false
		}
	}
	if len(o.TrustedIPs) != len(other.TrustedIPs) {
		return false
	}
	for i, ip := range o.TrustedIPs {
		if ip != other.TrustedIPs[i] {
			return false
		}
	}
	if len(o.EmailDomains) != len(other.EmailDomains) {
		return false
	}
	for i, domain := range o.EmailDomains {
		if domain != other.EmailDomains[i] {
			return false
		}
	}
	if len(o.WhitelistDomains) != len(other.WhitelistDomains) {
		return false
	}
	for i, domain := range o.WhitelistDomains {
		if domain != other.WhitelistDomains[i] {
			return false
		}
	}
	if len(o.HtpasswdUserGroups) != len(other.HtpasswdUserGroups) {
		return false
	}
	for i, group := range o.HtpasswdUserGroups {
		if group != other.HtpasswdUserGroups[i] {
			return false
		}
	}
	if o.ProxyPrefix != other.ProxyPrefix {
		return false
	}
	if o.PingPath != other.PingPath {
		return false
	}
	if o.PingUserAgent != other.PingUserAgent {
		return false
	}
	if o.ReadyPath != other.ReadyPath {
		return false
	}
	if o.RealClientIPHeader != other.RealClientIPHeader {
		return false
	}
	if o.RawRedirectURL != other.RawRedirectURL {
		return false
	}
	if o.AuthenticatedEmailsFile != other.AuthenticatedEmailsFile {
		return false
	}
	if o.HtpasswdFile != other.HtpasswdFile {
		return false
	}
	if !o.Cookie.isEqualTo(&other.Cookie) {
		return false
	}
	if !o.Session.isEqualTo(&other.Session) {
		return false
	}
	if !o.Templates.isEqualTo(&other.Templates) {
		return false
	}
	if !o.InjectRequestHeaders.isEqualTo(other.InjectRequestHeaders) {
		return false
	}
	if !o.InjectResponseHeaders.isEqualTo(other.InjectResponseHeaders) {
		return false
	}
	if len(o.Providers) != len(other.Providers) {
		return false
	}
	for i, provider := range o.Providers {
		if provider.ID != other.Providers[i].ID {
			return false
		}
		if provider.Type != other.Providers[i].Type {
			return false
		}
		if provider.Name != other.Providers[i].Name {
			return false
		}
		if provider.ClientID != other.Providers[i].ClientID {
			return false
		}
		if provider.ClientSecret != other.Providers[i].ClientSecret {
			return false
		}
		if provider.ClientSecretFile != other.Providers[i].ClientSecretFile {
			return false
		}
		if provider.Scope != other.Providers[i].Scope {
			return false
		}
		if provider.LoginURL != other.Providers[i].LoginURL {
			return false
		}
		if provider.RedeemURL != other.Providers[i].RedeemURL {
			return false
		}
		if provider.ProfileURL != other.Providers[i].ProfileURL {
			return false
		}
		if provider.ValidateURL != other.Providers[i].ValidateURL {
			return false
		}
		if provider.CodeChallengeMethod != other.Providers[i].CodeChallengeMethod {
			return false
		}
	}
	return true
}

func (c *ClaimSource) oauth2proxyClaimSource() options.ClaimSource {
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

func (h *Header) oauth2proxyHeader() *options.Header {
	return &options.Header{
		Name:                 h.Name,
		PreserveRequestValue: h.PreserveRequestValue,
		Values:               h.Values.oauth2proxyHeaderValues(),
	}
}

func (h *HeaderValue) oauth2proxyHeaderValue() options.HeaderValue {
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
		claimSource = h.ClaimSource.oauth2proxyClaimSource()
	}
	return options.HeaderValue{
		SecretSource: &source,
		ClaimSource:  &claimSource,
	}
}

func (h Headers) oauth2proxyHeaders() []options.Header {
	var headers []options.Header
	for _, header := range h {
		headers = append(headers, *header.oauth2proxyHeader())
	}
	return headers
}

func (h HeaderValues) oauth2proxyHeaderValues() []options.HeaderValue {
	var values []options.HeaderValue
	for _, hv := range h {
		values = append(values, hv.oauth2proxyHeaderValue())
	}
	return values
}

func (s SecretSource) isEqualTo(other *SecretSource) bool {
	if len(s.Value) != len(other.Value) {
		return false
	}
	for i, v := range s.Value {
		if v != other.Value[i] {
			return false
		}
	}
	if s.FromEnv != other.FromEnv {
		return false
	}
	if s.FromFile != other.FromFile {
		return false
	}
	return true
}

func (c *ClaimSource) isEqualTo(other *ClaimSource) bool {
	if c.Claim != other.Claim {
		return false
	}
	if c.Prefix != other.Prefix {
		return false
	}
	if c.BasicAuthPassword != nil && other.BasicAuthPassword != nil {
		return c.BasicAuthPassword.isEqualTo(other.BasicAuthPassword)
	}
	return true
}

func (v *HeaderValue) isEqualTo(other *HeaderValue) bool {
	if v.SecretSource != nil && other.SecretSource != nil {
		return v.SecretSource.isEqualTo(other.SecretSource)
	}
	if v.ClaimSource != nil && other.ClaimSource != nil {
		return v.ClaimSource.isEqualTo(other.ClaimSource)
	}
	return false
}

func (v HeaderValues) isEqualTo(other HeaderValues) bool {
	if len(v) != len(other) {
		return false
	}
	for i, v := range v {
		if !v.isEqualTo(&(other)[i]) {
			return false
		}
	}
	return true
}

func (h Header) isEqualTo(other Header) bool {
	if h.Name != other.Name {
		return false
	}
	if h.PreserveRequestValue != other.PreserveRequestValue {
		return false
	}
	return h.Values.isEqualTo(other.Values)
}

func (h Headers) isEqualTo(other Headers) bool {
	if len(h) != len(other) {
		return false
	}
	for i, v := range h {
		if !v.isEqualTo(other[i]) {
			return false
		}
	}
	return true
}

func (t *Templates) isEqualTo(other *Templates) bool {
	if t.Path != other.Path {
		return false
	}
	if t.CustomLogo != other.CustomLogo {
		return false
	}
	if t.Banner != other.Banner {
		return false
	}
	if t.Footer != other.Footer {
		return false
	}
	if t.DisplayLoginForm != other.DisplayLoginForm {
		return false
	}
	if t.Debug != other.Debug {
		return false
	}
	return true
}

func (c *CookieStoreOptions) isEqualTo(other *CookieStoreOptions) bool {
	return c.Minimal == other.Minimal
}

func (r *RedisStoreOptions) isEqualTo(other *RedisStoreOptions) bool {
	if r.ConnectionURL != other.ConnectionURL {
		return false
	}
	if r.Password != other.Password {
		return false
	}
	if r.UseSentinel != other.UseSentinel {
		return false
	}
	if r.SentinelPassword != other.SentinelPassword {
		return false
	}
	if r.SentinelMasterName != other.SentinelMasterName {
		return false
	}
	if r.SentinelConnectionURLs != nil && other.SentinelConnectionURLs != nil {
		if len(r.SentinelConnectionURLs) != len(other.SentinelConnectionURLs) {
			return false
		}
		for i, v := range r.SentinelConnectionURLs {
			if v != other.SentinelConnectionURLs[i] {
				return false
			}
		}
	}
	if r.UseCluster != other.UseCluster {
		return false
	}
	if r.ClusterConnectionURLs != nil || other.ClusterConnectionURLs != nil {
		if len(r.ClusterConnectionURLs) != len(other.ClusterConnectionURLs) {
			return false
		}
		for i, v := range r.ClusterConnectionURLs {
			if v != other.ClusterConnectionURLs[i] {
				return false
			}
		}
	}
	if r.CAPath != other.CAPath {
		return false
	}
	if r.InsecureSkipTLSVerify != other.InsecureSkipTLSVerify {
		return false
	}
	if r.IdleTimeout != other.IdleTimeout {
		return false
	}
	return true
}

func (s *SessionOptions) isEqualTo(other *SessionOptions) bool {
	if s.Type != other.Type {
		return false
	}
	if s.Cookie.isEqualTo(&other.Cookie) {
		return false
	}
	if s.Redis.isEqualTo(&other.Redis) {
		return false
	}
	return true
}

func (c *Cookie) isEqualTo(other *Cookie) bool {
	if c.Name != other.Name {
		return false
	}
	if c.Secret != other.Secret {
		return false
	}
	if c.Domains != nil && other.Domains != nil {
		if len(c.Domains) != len(other.Domains) {
			return false
		}
		for i, v := range c.Domains {
			if v != other.Domains[i] {
				return false
			}
		}
	}
	if c.Path != other.Path {
		return false
	}
	if c.Expire != other.Expire {
		return false
	}
	if c.Refresh != other.Refresh {
		return false
	}
	if c.NoSecure != other.NoSecure {
		return false
	}
	if c.NoHTTPOnly != other.NoHTTPOnly {
		return false
	}
	if c.SameSite != other.SameSite {
		return false
	}
	if c.CSRFPerRequest != other.CSRFPerRequest {
		return false
	}
	if c.CSRFExpire != other.CSRFExpire {
		return false
	}
	return true
}
