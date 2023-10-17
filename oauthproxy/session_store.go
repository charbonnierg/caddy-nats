package oauthproxy

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/sessions"
)

type SessionStore interface {
	Store() sessionsapi.SessionStore
	Provision(opts *options.Cookie) error
}

func (s *CookieStore) Store() sessionsapi.SessionStore { return s.store }

type CookieStore struct {
	store   sessionsapi.SessionStore
	Minimal bool `json:"session_cookie_minimal"`
}

func (s *CookieStore) Provision(opts *options.Cookie) error {
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
		ID:  "oauth2.session_store.cookie",
		New: func() caddy.Module { return new(CookieStore) },
	}
}

var (
	_ SessionStore = (*CookieStore)(nil)
)
