// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure_keyvault

import (
	"errors"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/pkg/azutils"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AzureKeyvault{})
}

// AzureKeyvault is a secrets secrets.Store implementation
// that retrieves secrets from Azure Keyvault.
type AzureKeyvault struct {
	ctx    caddy.Context
	logger *zap.Logger
	client AzureKeyvaultClient `json:"-"`
	// The Azure Keyvault URI
	URI string        `json:"uri,omitempty"`
	TTL time.Duration `json:"ttl,omitempty"`
	// The Azure Keyvault credential config
	CredentialConfig *azutils.CredentialConfig `json:"credential,omitempty"`
}

// CaddyModule returns the Caddy module information.
// It is required to implement the secrets.Store interface.
func (AzureKeyvault) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.stores.azure_keyvault",
		New: func() caddy.Module { return new(AzureKeyvault) },
	}
}

// Provision prepares the store for use.
// It is required to implement the secrets.Store interface.
func (s *AzureKeyvault) Provision(app secrets.App) error {
	if s.URI == "" {
		return errors.New("uri is required")
	}
	s.ctx = app.Context()
	s.logger = s.ctx.Logger().Named("store.azure_keyvault").With(zap.String("uri", s.URI))
	if s.CredentialConfig == nil {
		s.CredentialConfig = new(azutils.CredentialConfig)
	}
	err := s.CredentialConfig.ParseEnv().Build()
	if err != nil {
		return err
	}
	if s.TTL == 0 {
		s.TTL = 1 * time.Minute
	}
	s.logger.Info("provisioning azure keyvault store")
	client, err := NewClient(s.URI, s.TTL, s.CredentialConfig)
	if err != nil {
		s.logger.Error("error creating azure keyvault client", zap.Error(err))
		return err
	}
	s.client = client
	return nil
}

// Get retrieves a value from the store for a given key.
// It is required to implement the secrets.Store interface.
func (s *AzureKeyvault) Get(key string) (string, error) {
	s.logger.Info("getting secret", zap.String("key", key))
	v, err := s.client.GetSecret(s.ctx, key)
	if err != nil {
		s.logger.Error("error getting secret", zap.Error(err))
	}
	return v, err
}

// Interface guards
var (
	_ secrets.Store         = (*AzureKeyvault)(nil)
	_ caddyfile.Unmarshaler = (*AzureKeyvault)(nil)
)
