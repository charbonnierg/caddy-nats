package session_store

import (
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	nats "github.com/charbonnierg/beyond/modules/nats/interfaces"
	"github.com/charbonnierg/beyond/modules/oauth2"
	"github.com/charbonnierg/beyond/modules/oauth2/session_store/jetstream"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamStore{})
}

type JetStreamStore struct {
	logger        *zap.Logger
	sessionsstore sessionsapi.SessionStore
	kvstore       jetstream.Store
	Name          string            `json:"name,omitempty"`
	Client        *jetstream.Client `json:"client,omitempty"`
	TTL           time.Duration     `json:"ttl,omitempty"`
}

func (s *JetStreamStore) Provision(app *oauth2.App, opts *options.Cookie) error {
	s.logger, _ = zap.NewDevelopment()
	if s.Client.Internal {
		unm, err := app.LoadApp("nats")
		if err != nil {
			return err
		}
		natsApp, ok := unm.(nats.NatsApp)
		if !ok {
			return errors.New("invalid nats app module")
		}
		s.Client.SetInProcessServerProvider(natsApp)
	}
	jsstore := jetstream.NewStore(s.Name, s.Client, s.TTL, s.logger)
	s.kvstore = *jsstore
	store, err := jsstore.SessionStore(opts)
	if err != nil {
		return fmt.Errorf("failed to create jetstream session store: %v", err)
	}
	s.sessionsstore = store
	return nil
}

func (JetStreamStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauth2.session_store.jetstream",
		New: func() caddy.Module { return new(JetStreamStore) },
	}
}

func (s *JetStreamStore) Store() sessionsapi.SessionStore {
	return s.sessionsstore
}

var (
	_ oauth2.SessionStore = (*JetStreamStore)(nil)
)
