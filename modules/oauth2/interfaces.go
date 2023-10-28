// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/providers/util"
	"github.com/quara-dev/beyond"
)

type App interface {
	beyond.App
	beyond.BeyondAppLoader
	GetEndpoint(name string) (Endpoint, error)
	GetOrAddEndpoint(endpoint Endpoint) (Endpoint, error)
	GetReplacer() *caddy.Replacer
}

type Endpoint interface {
	Provision(app App) error
	Name() string
	Setup() error
	Equals(other Endpoint) bool
	IsReference() bool
	ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error
	DecodeSessionState(session []*http.Cookie) (*sessions.SessionState, error)
	DecodeSessionStateFromString(session string) (*sessions.SessionState, error)
	GetOidcSessionClaimExtractor(session *sessions.SessionState) (util.ClaimExtractor, error)
}

type Store interface {
	caddy.Module
	GetStore() sessions.SessionStore
	Provision(app App, opts *options.Cookie) error
}
