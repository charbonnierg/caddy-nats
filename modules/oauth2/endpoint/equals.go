// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package endpoint

// Return true if two Options structs are equal.
func (o *Options) equals(other *Options) bool {
	if o == nil && other == nil {
		return true
	}
	if o == nil && other != nil {
		return false
	}
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
	if !o.Cookie.equals(other.Cookie) {
		return false
	}
	if !o.Templates.equals(other.Templates) {
		return false
	}
	if !o.InjectRequestHeaders.equals(other.InjectRequestHeaders) {
		return false
	}
	if !o.InjectResponseHeaders.equals(other.InjectResponseHeaders) {
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

// Return true if two SecretSource structs are equal.
func (s SecretSource) equals(other *SecretSource) bool {
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

// Return true if two ClaimSource structs are equal.
func (c *ClaimSource) equals(other *ClaimSource) bool {
	if c.Claim != other.Claim {
		return false
	}
	if c.Prefix != other.Prefix {
		return false
	}
	if c.BasicAuthPassword != nil && other.BasicAuthPassword != nil {
		return c.BasicAuthPassword.equals(other.BasicAuthPassword)
	}
	return true
}

// Return true if two HeaderValue structs are equal.
func (v *HeaderValue) equals(other *HeaderValue) bool {
	if v.SecretSource != nil && other.SecretSource != nil {
		return v.SecretSource.equals(other.SecretSource)
	}
	if v.ClaimSource != nil && other.ClaimSource != nil {
		return v.ClaimSource.equals(other.ClaimSource)
	}
	return false
}

// Return true if two HeaderValues structs are equal.
func (v HeaderValues) equals(other HeaderValues) bool {
	if len(v) != len(other) {
		return false
	}
	for i, v := range v {
		if !v.equals(&(other)[i]) {
			return false
		}
	}
	return true
}

// Return true if two Header structs are equal.
func (h Header) equals(other Header) bool {
	if h.Name != other.Name {
		return false
	}
	if h.PreserveRequestValue != other.PreserveRequestValue {
		return false
	}
	return h.Values.equals(other.Values)
}

// Return true if two Headers structs are equal.
func (h Headers) equals(other Headers) bool {
	if len(h) != len(other) {
		return false
	}
	for i, v := range h {
		if !v.equals(other[i]) {
			return false
		}
	}
	return true
}

// Return true if two Templates structs are equal.
func (t *Templates) equals(other *Templates) bool {
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

// Return true if two Cookie structs are equal.
func (c *Cookie) equals(other *Cookie) bool {
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
