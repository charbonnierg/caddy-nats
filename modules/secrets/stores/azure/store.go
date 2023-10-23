package azure

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond/modules/secrets"
)

type AzureKeyvault struct {
	ctx    caddy.Context
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
	client, err := NewAzureKeyvaultClient(s.URI, s.CredentialConfig)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

// Get retrieves a value from the store for a given key.
func (s *AzureKeyvault) Get(key string) (string, error) {
	return s.client.GetSecret(s.ctx, key)
}

// Set writes a value to the store for a given existing key.
func (s *AzureKeyvault) Set(key string, value string) error {
	return s.client.SetSecret(s.ctx, key, value)
}

// Interface guards
var (
	_ secrets.Store = (*AzureKeyvault)(nil)
)
