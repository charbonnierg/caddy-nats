// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cookie

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/sessions"
	"github.com/quara-dev/beyond/modules/oauth2"
)

func init() {
	caddy.RegisterModule(CookieStore{})
}

type CookieStore struct {
	store   sessionsapi.SessionStore
	Minimal bool `json:"session_cookie_minimal"`
}

func (s *CookieStore) GetStore() sessionsapi.SessionStore { return s.store }

func (s *CookieStore) Provision(app oauth2.App, opts *options.Cookie) error {
	storeOpts := &options.SessionOptions{Type: options.CookieSessionStoreType, Cookie: options.CookieStoreOptions{Minimal: s.Minimal}}
	store, err := sessions.NewSessionStore(storeOpts, opts)
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

func (CookieStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  oauth2.StoreID("cookie"),
		New: func() caddy.Module { return new(CookieStore) },
	}
}

var (
	_ oauth2.Store = (*CookieStore)(nil)
)
