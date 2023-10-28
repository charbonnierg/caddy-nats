// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream

import (
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/oauth2"
	"github.com/quara-dev/beyond/modules/oauth2/stores/jetstream/internal"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamStore{})
}

type JetStreamStore struct {
	logger        *zap.Logger
	sessionsstore sessionsapi.SessionStore
	Name          string            `json:"name,omitempty"`
	Client        *natsutils.Client `json:"client,omitempty"`
	TTL           time.Duration     `json:"ttl,omitempty"`
}

func (s *JetStreamStore) Provision(app oauth2.App, opts *options.Cookie) error {
	s.logger, _ = zap.NewDevelopment()
	if s.Client.Internal {
		unm, err := app.LoadBeyondApp("nats")
		if err != nil {
			return err
		}
		natsApp, ok := unm.(nats.NatsApp)
		if !ok {
			return errors.New("invalid nats app module")
		}
		s.Client.SetInternalProvider(natsApp)
	}
	store, err := internal.NewStore(s.Name, s.Client, s.TTL, s.logger).SessionStore(opts)
	if err != nil {
		return fmt.Errorf("failed to create jetstream session store: %v", err)
	}
	s.sessionsstore = store
	return nil
}

func (JetStreamStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  oauth2.StoreID("jetstream"),
		New: func() caddy.Module { return new(JetStreamStore) },
	}
}

func (s *JetStreamStore) GetStore() sessionsapi.SessionStore {
	return s.sessionsstore
}

var (
	_ oauth2.Store = (*JetStreamStore)(nil)
)
