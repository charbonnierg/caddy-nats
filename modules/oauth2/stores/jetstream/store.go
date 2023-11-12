// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream

import (
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/modules/oauth2"
	"github.com/quara-dev/beyond/modules/oauth2/stores/jetstream/internal"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamStore{})
}

type JetStreamStore struct {
	ctx           caddy.Context
	logger        *zap.Logger
	sessionsstore sessionsapi.SessionStore
	Name          string                 `json:"name,omitempty"`
	Account       string                 `json:"account,omitempty"`
	Client        *natsclient.NatsClient `json:"connection,omitempty"`
	TTL           time.Duration          `json:"ttl,omitempty"`
}

func (s *JetStreamStore) Provision(app oauth2.App, opts *options.Cookie) error {
	s.ctx = app.Context()
	s.logger = app.Logger().Named("oauth2-jetstream")
	if s.Client == nil {
		s.Client = &natsclient.NatsClient{Internal: true}
	}
	if err := caddynats.ProvisionClientConnectionBeyond(s.ctx, s.Account, s.Client); err != nil {
		return err
	}
	store, err := internal.NewStore(s.ctx, s.Name, s.Client, s.TTL, s.logger).SessionStore(opts)
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
	s.logger.Info("Loading store", zap.String("name", s.Name))
	return s.sessionsstore
}

var (
	_ oauth2.Store = (*JetStreamStore)(nil)
)
