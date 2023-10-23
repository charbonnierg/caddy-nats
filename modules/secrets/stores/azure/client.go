package azure

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

func NewAzureKeyvaultClient(uri string, cfg *AzCredentialConfig) (*AzureKeyvaultClient, error) {
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

type AzureKeyvaultClient struct {
	cfg    *AzCredentialConfig
	client *azsecrets.Client
	URI    string
}

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

func (c *AzureKeyvaultClient) SetSecret(ctx context.Context, name string, value string) error {
	v := string(value)
	params := azsecrets.SetSecretParameters{Value: &v}
	_, err := c.client.SetSecret(ctx, name, params, nil)
	return err
}
