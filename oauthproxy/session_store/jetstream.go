package session_store

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/caddy-nats/oauthproxy"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/sessions"
)

func (s *JetStreamStore) Store() sessionsapi.SessionStore { return s.store }

type JetStreamStore struct {
	store sessionsapi.SessionStore
}

func (s *JetStreamStore) Provision(opts *options.Cookie) error {
	storeOpts := &options.SessionOptions{}
	store, err := sessions.NewSessionStore(storeOpts, opts)
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

func (JetStreamStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauth2.session_store.jetstream",
		New: func() caddy.Module { return new(JetStreamStore) },
	}
}

var (
	_ oauthproxy.SessionStore = (*JetStreamStore)(nil)
)
