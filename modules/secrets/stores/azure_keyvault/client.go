// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure_keyvault

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/quara-dev/beyond/pkg/azutils"
)

type AzureKeyvaultClient interface {
	GetSecret(ctx context.Context, name string) (string, error)
}

// NewClient creates a new AzureKeyvaultClient
func NewClient(uri string, cfg *azutils.CredentialConfig) (AzureKeyvaultClient, error) {
	creds, err := cfg.Authenticate()
	if err != nil {
		return nil, err
	}
	client, err := azsecrets.NewClient(uri, creds, nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

// Client is a wrapper around azsecrets.Client which implements AazureKeyvaultClient
type Client struct {
	client *azsecrets.Client
	URI    string
}

// GetSecret gets a secret from Azure Keyvault
func (c *Client) GetSecret(ctx context.Context, name string) (string, error) {
	resp, err := c.client.GetSecret(ctx, name, "", nil)
	if err != nil {
		return "", err
	}
	if resp.Value == nil {
		return "", errors.New("secret value is nil")
	}
	return *resp.Value, nil
}
