// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure_keyvault

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/quara-dev/beyond/pkg/azutils"
)

// AzureKeyvaultClient is a wrapper around azsecrets.Client
type AzureKeyvaultClient struct {
	cfg    *azutils.CredentialConfig
	client *azsecrets.Client
	URI    string
}

// GetSecret gets a secret from Azure Keyvault
func (c *AzureKeyvaultClient) GetSecret(ctx context.Context, name string) (string, error) {
	resp, err := c.client.GetSecret(ctx, name, "", nil)
	if err != nil {
		return "", err
	}
	if resp.Value == nil {
		return "", errors.New("secret value is nil")
	}
	return *resp.Value, nil
}

// NewAzureKeyvaultClient creates a new AzureKeyvaultClient
func NewAzureKeyvaultClient(uri string, cfg *azutils.CredentialConfig) (*AzureKeyvaultClient, error) {
	creds, err := cfg.NewCredential()
	if err != nil {
		return nil, err
	}
	client, err := azsecrets.NewClient(uri, creds, nil)
	if err != nil {
		return nil, err
	}
	return &AzureKeyvaultClient{
		cfg:    cfg,
		client: client,
	}, nil
}
