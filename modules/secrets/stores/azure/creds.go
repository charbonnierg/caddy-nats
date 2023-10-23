// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package azure

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

var (
	ErrTenantIdNotSpecified     = errors.New("tenant_id or tenant_id_file must be specified")
	ErrClientIdNotSpecified     = errors.New("client_id or client_id_file must be specified")
	ErrClientSecretNotSpecified = errors.New("client_secret or client_secret_file must be specified")
)

// NewAzCredentialConfig creates a new AzCredentialConfig with all fields set to their zero value.
// Environment variables are NOT parsed. If you want to parse environment
// variables, use ParseEnv method after creating the config.
// This method is strictly equivalent to:
//
//	&AzCredentialConfig{}
func NewAzCredentialConfig() *AzCredentialConfig {
	return &AzCredentialConfig{}
}

// AzCredentialConfig is a configuration for creating an Azure credential.
// All fields are optional.
// Basic usage:
//
//	// Create a config from env
//	azc := azure.NewAzCredentialConfig().ParseEnv()
//	// Validate config (optional)
//	err := azc.ParseEnv().Build()
//	if err != nil {
//	// Handle error
//	// ...
//	}
//	// Create a credential
//	cred, err := azc.NewCredential()
//	if err != nil {
//	// Handle error
//	// ...
//	}
//	// Use the credential
//	// ...
type AzCredentialConfig struct {
	// factory is the generated factory method for creating a credential according to the config.
	factory func() (azcore.TokenCredential, error)
	// Public fields are used for JSON unmarshalling.
	ClientId                   string   `json:"client_id,omitempty"`
	ClientIdFile               string   `json:"client_id_file,omitempty"`
	ClientSecret               string   `json:"client_secret,omitempty"`
	ClientSecretFile           string   `json:"client_secret_file,omitempty"`
	TenantId                   string   `json:"tenant_id,omitempty"`
	TenantIdFile               string   `json:"tenant_id_file,omitempty"`
	NoDefaultCredentials       bool     `json:"no_default_credentials,omitempty"`
	NoManagedIdentity          bool     `json:"no_managed_identity,omitempty"`
	ManagedIdentityClientId    string   `json:"managed_identity_client_id,omitempty"`
	ManagedIdentityResourceId  string   `json:"managed_identity_resource_id,omitempty"`
	AdditionallyAllowedTenants []string `json:"additionally_allowed_tenants,omitempty"`
	DisableInstanceDiscovery   bool     `json:"disable_instance_discovery,omitempty"`
	// These are for testing purposes only
	clientSecretCredentialFactory    func(tenantID string, clientID string, clientSecret string, options *azidentity.ClientSecretCredentialOptions) (*azidentity.ClientSecretCredential, error)
	managedIdentityCredentialFactory func(options *azidentity.ManagedIdentityCredentialOptions) (*azidentity.ManagedIdentityCredential, error)
	defaultAzureCredentialFactory    func(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error)
}

// ParseEnv parses the environment variables for Azure credentials.
// It will override any existing field in the config if a non-empty
// value exists as environment variable.
// The following environment variables are supported (all variables are optional):
//   - AZURE_CLIENT_ID: the client ID of a service principal
//   - AZURE_CLIENT_ID_FILE: the path to a file containing the client ID of a service principal
//   - AZURE_CLIENT_SECRET: the client secret of a service principal
//   - AZURE_CLIENT_SECRET_FILE: the path to a file containing the client secret of a service principal
//   - AZURE_TENANT_ID: the tenant ID of the service principal
//   - AZURE_TENANT_ID_FILE: the path to a file containing the tenant ID of the service principal
//   - NO_DEFAULT_CREDENTIALS: if set to true, the default credentials will not be used. Default credentials are used when no client id or managed identity is specified.
//   - NO_MANAGED_IDENTITY: if set to true, managed identity will not be used. Managed identity is used when no client id is specified.
//   - DISABLE_INSTANCE_DISCOVERY: if set true skip request for Azure AD instance metadata from https://login.microsoft.com before authenticating, making the application responsible for ensuring the configured authority is valid and trustworthy.
//     This should only be set to true by applications authenticating in disconnected clouds, or private clouds such as Azure Stack.
//   - MANAGED_IDENTITY_CLIENT_ID: the client ID of a user-assigned managed identity
//   - MANAGED_IDENTITY_RESOURCE_ID: the resource ID of a user-assigned managed identity
//   - ADDITIONALLY_ALLOWED_TENANTS: a comma-separated list of tenant IDs that are additionally allowed to authenticate
func (azc *AzCredentialConfig) ParseEnv() *AzCredentialConfig {
	clientId, ok := os.LookupEnv("AZURE_CLIENT_ID")
	if ok && clientId != "" {
		azc.ClientId = clientId
	}
	clientIdFile, ok := os.LookupEnv("AZURE_CLIENT_ID_FILE")
	if ok && clientIdFile != "" {
		azc.ClientIdFile = clientIdFile
	}
	clientSecret, ok := os.LookupEnv("AZURE_CLIENT_SECRET")
	if ok && clientSecret != "" {
		azc.ClientSecret = clientSecret
	}
	clientSecretFile, ok := os.LookupEnv("AZURE_CLIENT_SECRET_FILE")
	if ok && clientSecretFile != "" {
		azc.ClientSecretFile = clientSecretFile
	}
	tenantId, ok := os.LookupEnv("AZURE_TENANT_ID")
	if ok && tenantId != "" {
		azc.TenantId = tenantId
	}
	tenantIdFile, ok := os.LookupEnv("AZURE_TENANT_ID_FILE")
	if ok && tenantIdFile != "" {
		azc.TenantIdFile = tenantIdFile
	}
	noDefaultCredentials, ok := os.LookupEnv("NO_DEFAULT_CREDENTIALS")
	if ok && noDefaultCredentials != "" {
		azc.NoDefaultCredentials, _ = strconv.ParseBool(noDefaultCredentials)
	}
	noManagedIdentity, ok := os.LookupEnv("NO_MANAGED_IDENTITY")
	if ok && noManagedIdentity != "" {
		azc.NoManagedIdentity, _ = strconv.ParseBool(noManagedIdentity)
	}
	disableInstanceDiscovery, ok := os.LookupEnv("DISABLE_INSTANCE_DISCOVERY")
	if ok && disableInstanceDiscovery != "" {
		azc.DisableInstanceDiscovery, _ = strconv.ParseBool(disableInstanceDiscovery)
	}
	managedIdentityClientId, ok := os.LookupEnv("MANAGED_IDENTITY_CLIENT_ID")
	if ok && managedIdentityClientId != "" {
		azc.ManagedIdentityClientId = managedIdentityClientId
	}
	managedIdentityResourceId, ok := os.LookupEnv("MANAGED_IDENTITY_RESOURCE_ID")
	if ok && managedIdentityResourceId != "" {
		azc.ManagedIdentityResourceId = managedIdentityResourceId
	}
	additionallyAllowedTenants, ok := os.LookupEnv("ADDITIONALLY_ALLOWED_TENANTS")
	if ok && additionallyAllowedTenants != "" {
		azc.AdditionallyAllowedTenants = strings.Split(additionallyAllowedTenants, ",")
	}
	return azc
}

// GetTenantId returns the tenant ID specified in the config if it exists.
// If the tenant ID is not specified, it returns an error (ErrTenantIdNotSpecified).
// If the tenant ID cannot be read from the config, it returns a wrapped error.
func (azc *AzCredentialConfig) GetTenantId() (string, error) {
	switch {
	case azc.TenantId != "":
		if azc.TenantIdFile != "" {
			return "", errors.New("tenant_id and tenant_id_file are mutually exclusive")
		}
		return azc.TenantId, nil
	case azc.TenantIdFile != "":
		content, err := os.ReadFile(azc.TenantIdFile)
		if err != nil {
			return "", fmt.Errorf("failed to read tenant_id_file: %v", err)
		}
		return string(content), nil
	default:
		return "", ErrTenantIdNotSpecified
	}
}

// GetClientId returns the client ID specified in the config if it exists.
// If the client ID is not specified, it returns an error (ErrClientIdNotSpecified).
// If the client ID cannot be read from the config, it returns a wrapped error.
func (azc *AzCredentialConfig) GetClientId() (string, error) {
	switch {
	case azc.ClientId != "":
		if azc.ClientIdFile != "" {
			return "", errors.New("client_id and client_id_file are mutually exclusive")
		}
		return azc.ClientId, nil
	case azc.ClientIdFile != "":
		content, err := os.ReadFile(azc.ClientIdFile)
		if err != nil {
			return "", fmt.Errorf("failed to read client_id_file: %v", err)
		}
		return string(content), nil
	default:
		return "", ErrClientIdNotSpecified
	}
}

// GetClientSecret returns the client secret specified in the config if it exists.
// If the client secret is not specified, it returns an error (ErrClientSecretNotSpecified).
// If the client secret cannot be read from the config, it returns a wrapped error.
func (azc *AzCredentialConfig) GetClientSecret() (string, error) {
	switch {
	case azc.ClientSecret != "":
		if azc.ClientSecretFile != "" {
			return "", errors.New("client_secret and client_secret_file are mutually exclusive")
		}
		return azc.ClientSecret, nil
	case azc.ClientSecretFile != "":
		content, err := os.ReadFile(azc.ClientSecretFile)
		if err != nil {
			return "", fmt.Errorf("failed to read client_secret_file: %v", err)
		}
		return string(content), nil
	default:
		return "", ErrClientSecretNotSpecified
	}
}

// Build builds the credential factory according to the config.
// If the config is invalid, it returns an error.
// This method is called automatically by NewCredential when required so you don't need
// to call it manually.
func (azc *AzCredentialConfig) Build() error {
	switch {
	case azc.ClientSecret != "" || azc.ClientSecretFile != "":
		clientSecret, err := azc.GetClientSecret()
		if err != nil {
			return err
		}
		clientId, err := azc.GetClientId()
		if err != nil {
			return err
		}
		tenant, err := azc.GetTenantId()
		if err != nil {
			return err
		}
		azc.factory = func() (azcore.TokenCredential, error) {
			return azc.newClientSecretCredential(tenant, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{
				AdditionallyAllowedTenants: azc.AdditionallyAllowedTenants,
				DisableInstanceDiscovery:   azc.DisableInstanceDiscovery,
			})
		}
		return nil
	case azc.NoManagedIdentity && azc.NoDefaultCredentials:
		return errors.New("no credentials specified")
	case azc.NoDefaultCredentials && azc.ManagedIdentityClientId != "":
		if azc.ManagedIdentityResourceId != "" {
			return errors.New("managed_identity_client_id and managed_identity_resource_id are mutually exclusive")
		}
		if azc.ClientId != "" || azc.ClientIdFile != "" {
			return errors.New("managed_identity_client_id and client_id are mutually exclusive")
		}
		azc.factory = func() (azcore.TokenCredential, error) {
			return azc.newManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(azc.ManagedIdentityClientId),
			})
		}
		return nil
	case azc.NoDefaultCredentials && azc.ManagedIdentityResourceId != "":
		if azc.ClientId != "" || azc.ClientIdFile != "" {
			return errors.New("managed_identity_resource_id and client_id are mutually exclusive")
		}
		azc.factory = func() (azcore.TokenCredential, error) {
			return azc.newManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ResourceID(azc.ManagedIdentityResourceId),
			})
		}
		return nil
	case azc.NoDefaultCredentials:
		azc.factory = func() (azcore.TokenCredential, error) {
			return azc.newManagedIdentityCredential(nil)
		}
		return nil
	default:
		tenant, err := azc.GetTenantId()
		if err != nil && err != ErrTenantIdNotSpecified {
			return err
		}
		azc.factory = func() (azcore.TokenCredential, error) {
			return azc.newDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
				AdditionallyAllowedTenants: azc.AdditionallyAllowedTenants,
				DisableInstanceDiscovery:   azc.DisableInstanceDiscovery,
				TenantID:                   tenant,
			})
		}
		return nil
	}
}

// NewCredential creates a new Azure credential based on the config.
// If the config is invalid, it returns an error.
func (azc *AzCredentialConfig) NewCredential() (azcore.TokenCredential, error) {
	if azc.factory == nil {
		err := azc.Build()
		if err != nil {
			return nil, err
		}
	}
	return azc.factory()
}

// SetCredentialFactories sets the factories for creating credentials.
// If a factory is nil, the default factory will be used.
// This method is mainly useful for testing. Don't use it in your application unless you want
// to customize the credential creation process.
func (azc *AzCredentialConfig) SetCredentialFactories(factories CredentialFactories) {
	if factories.clientSecretCredential != nil {
		azc.clientSecretCredentialFactory = factories.clientSecretCredential
	}
	if factories.managedIdentityCredential != nil {
		azc.managedIdentityCredentialFactory = factories.managedIdentityCredential
	}
	if factories.defaultAzureCredential != nil {
		azc.defaultAzureCredentialFactory = factories.defaultAzureCredential
	}
}

// Default azure credential wrapper for azidentity package
func (azc *AzCredentialConfig) newDefaultAzureCredential(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error) {
	factory := azc.defaultAzureCredentialFactory
	if factory == nil {
		factory = azidentity.NewDefaultAzureCredential
	}
	return factory(options)
}

// Managed identity credential wrapper for azidentity package
func (azc *AzCredentialConfig) newManagedIdentityCredential(options *azidentity.ManagedIdentityCredentialOptions) (*azidentity.ManagedIdentityCredential, error) {
	factory := azc.managedIdentityCredentialFactory
	if factory == nil {
		factory = azidentity.NewManagedIdentityCredential
	}
	return factory(options)
}

// Client secret credential wrapper for azidentity package
func (azc *AzCredentialConfig) newClientSecretCredential(tenantID string, clientID string, clientSecret string, options *azidentity.ClientSecretCredentialOptions) (*azidentity.ClientSecretCredential, error) {
	factory := azc.clientSecretCredentialFactory
	if factory == nil {
		factory = azidentity.NewClientSecretCredential
	}
	return factory(tenantID, clientID, clientSecret, options)
}

// CredentialFactories contains factories for creating credentials.
// If a factory is nil, the default factory will be used.
// This struct is mainly useful for testing. Don't use it in your application unless you want
// to customize the credential creation process.
type CredentialFactories struct {
	clientSecretCredential    func(tenantID string, clientID string, clientSecret string, options *azidentity.ClientSecretCredentialOptions) (*azidentity.ClientSecretCredential, error)
	managedIdentityCredential func(options *azidentity.ManagedIdentityCredentialOptions) (*azidentity.ManagedIdentityCredential, error)
	defaultAzureCredential    func(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error)
}
