// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AzureKeyvault{})
}

type AzureKeyvault struct {
	ctx    caddy.Context
	logger *zap.Logger
	client *AzureKeyvaultClient
	// The Azure Keyvault URI
	URI string `json:"uri,omitempty"`
	// The Azure Keyvault credential config
	CredentialConfig *AzCredentialConfig `json:"credential,omitempty"`
}

func (AzureKeyvault) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.store.azure_keyvault",
		New: func() caddy.Module { return new(AzureKeyvault) },
	}
}

// Provision prepares the store for use.
func (s *AzureKeyvault) Provision(app secrets.SecretApp) error {
	s.ctx = app.Context()
	s.logger = s.ctx.Logger().Named("azure_keyvault")
	if s.URI == "" {
		return errors.New("uri is required")
	}
	if s.CredentialConfig == nil {
		s.CredentialConfig = NewAzCredentialConfig()
	}
	err := s.CredentialConfig.ParseEnv().Build()
	if err != nil {
		return err
	}
	s.logger.Info("provisioning azure keyvault store", zap.String("uri", s.URI))
	client, err := NewAzureKeyvaultClient(s.URI, s.CredentialConfig)
	if err != nil {
		s.logger.Error("error creating azure keyvault client", zap.Error(err))
		return err
	}
	s.client = client
	return nil
}

// Get retrieves a value from the store for a given key.
func (s *AzureKeyvault) Get(key string) (string, error) {
	s.logger.Info("getting secret", zap.String("key", key))
	v, err := s.client.GetSecret(s.ctx, key)
	if err != nil {
		s.logger.Error("error getting secret", zap.Error(err))
	}
	return v, err
}

// Set writes a value to the store for a given existing key.
func (s *AzureKeyvault) Set(key string, value string) error {
	return s.client.SetSecret(s.ctx, key, value)
}

// Interface guards
var (
	_ secrets.Store         = (*AzureKeyvault)(nil)
	_ caddyfile.Unmarshaler = (*AzureKeyvault)(nil)
)
