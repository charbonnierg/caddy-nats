// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package session_store

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	sessionsapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/sessions"
	"github.com/quara-dev/beyond/modules/oauth2/oauth2app"
)

func init() {
	caddy.RegisterModule(RedisStore{})
}

type RedisStore struct {
	store                  sessionsapi.SessionStore
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

func (s *RedisStore) Store() sessionsapi.SessionStore { return s.store }

func (s *RedisStore) Provision(_ *oauth2app.App, opts *options.Cookie) error {
	storeOpts := &options.SessionOptions{
		Type: options.RedisSessionStoreType,
		Redis: options.RedisStoreOptions{
			ConnectionURL:         s.ConnectionURL,
			Password:              s.Password,
			UseSentinel:           s.UseSentinel,
			SentinelPassword:      s.SentinelPassword,
			SentinelMasterName:    s.SentinelMasterName,
			UseCluster:            s.UseCluster,
			ClusterConnectionURLs: s.ClusterConnectionURLs,
			CAPath:                s.CAPath,
			InsecureSkipTLSVerify: s.InsecureSkipTLSVerify,
			IdleTimeout:           s.IdleTimeout,
		},
	}
	store, err := sessions.NewSessionStore(storeOpts, opts)
	if err != nil {
		return err
	}
	s.store = store
	return nil
}

func (RedisStore) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "oauth2.session_store.redis",
		New: func() caddy.Module { return new(RedisStore) },
	}
}

var (
	_ oauth2app.SessionStore = (*RedisStore)(nil)
)
