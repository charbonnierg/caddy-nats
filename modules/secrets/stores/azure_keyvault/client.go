// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure_keyvault

import (
	"context"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/quara-dev/beyond/pkg/azutils"
)

type AzureKeyvaultClient interface {
	GetSecret(ctx context.Context, name string) (string, error)
}

// NewClient creates a new AzureKeyvaultClient
func NewClient(uri string, ttl time.Duration, cfg *azutils.CredentialConfig) (AzureKeyvaultClient, error) {
	creds, err := cfg.Authenticate()
	if err != nil {
		return nil, err
	}
	azclient, err := azsecrets.NewClient(uri, creds, nil)
	if err != nil {
		return nil, err
	}
	return &clientWrapper{
		client: azclient,
		cache:  make(map[string]*downloadedSecret),
		ttl:    ttl,
	}, nil
}

type downloadedSecret struct {
	value   string
	expires time.Time
}

func (s *downloadedSecret) Expired() bool {
	return time.Now().After(s.expires)
}

// Client is a wrapper around azsecrets.Client which implements AazureKeyvaultClient
type clientWrapper struct {
	client *azsecrets.Client
	cache  map[string]*downloadedSecret
	ttl    time.Duration
}

// GetSecret gets a secret from Azure Keyvault
func (c *clientWrapper) GetSecret(ctx context.Context, name string) (string, error) {
	if secret, ok := c.cache[name]; ok {
		if !secret.Expired() {
			return secret.value, nil
		}
	}
	resp, err := c.client.GetSecret(ctx, name, "", nil)
	if err != nil {
		return "", err
	}
	if resp.Value == nil {
		return "", errors.New("secret value is nil")
	}
	c.cache[name] = &downloadedSecret{
		value:   *resp.Value,
		expires: time.Now().Add(c.ttl),
	}
	return *resp.Value, nil
}
