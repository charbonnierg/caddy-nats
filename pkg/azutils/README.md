# `azutils`

> A library to help writing modules connecting to Azure services.

## Introduction

When writing modules relying on Azure services, it is always required to obtain credentials before interacting
with the services.

This library offers a simple struct to help obtaining credentials. It is based on the [azidentity](https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity#azure-identity-client-module-for-go) library.

## API

This module defines a single struct named `CredentialConfig`. It holds azure credential configuration and all fields are optional. Configuration can be edited after struct is created:

```go
import "github.com/quara-dev/beyond/pkg/azutils"

// Create a new config. An empty config is valid.
cfg := new(azutils.CredentialConfig)

// Set client id
cfg.ClientId = "<value>"
// Set client id file
cfg.ClientIdFile = "<value>"
// Set client secret
cfg.ClientSecret = "<value>"
// Set client secret file
cfg.ClientSecretFile = "<value>"
// Set tenant id
cfg.TenantId = "<value>"
// Set tenant id file
cfg.TenantIdFile = "<value>"
// Disable default credentials
cfg.NoDefaultCredentials = true
// Disable managed identities
cfg.NoManagedIdentity = true
// Set managed identity client id
cfg.ManagedIdentityClientId = "<value>"
// Set managed identity resource id
cfg.ManagedIdentityResourceId = "<value>"
// Set additionally allowed tenants by id
cfg.AdditionallyAllowedTenants = []string{"<value1>", "<value2>"}
// Disable instance discovery
cfg.DisableInstanceDiscovery = true
```

### Environment Variables

By default, environment variables are not used.

In order to set credential config fields according to environment variables, the `.ParseEnv()` method must be used:

```go
cfg := new(azutils.CredentialConfig).ParseEnv()
```

The following environment variables are supported (all variables are optional):
  - `AZURE_CLIENT_ID`: the client ID of a service principal
  - `AZURE_CLIENT_ID_FILE`: the path to a file containing the client ID of a service principal
  - `AZURE_CLIENT_SECRET`: the client secret of a service principal
  - `AZURE_CLIENT_SECRET_FILE`: the path to a file containing the client secret of a service principal
  - `AZURE_TENANT_ID`: the tenant ID of the service principal
  - `AZURE_TENANT_ID_FILE`: the path to a file containing the tenant ID of the service principal
  - `NO_DEFAULT_CREDENTIALS`: if set to true, the default credentials will not be used. Default credentials are used when no client id or managed identity is specified.
  - `NO_MANAGED_IDENTITY`: if set to true, managed identity will not be used. Managed identity is used when no client id is specified.
  - `DISABLE_INSTANCE_DISCOVERY`: if set true skip request for Azure AD instance metadata from https://login.microsoft.com before authenticating, making the application responsible for ensuring the configured authority is valid and trustworthy.
     This should only be set to true by applications authenticating in disconnected clouds, or private clouds such as Azure Stack.
  - MANAGED_IDENTITY_CLIENT_ID: the client ID of a user-assigned managed identity
  - `MANAGED_IDENTITY_RESOURCE_ID`: the resource ID of a user-assigned managed identity
  - `ADDITIONALLY_ALLOWED_TENANTS`: a comma-separated list of tenant IDs that are additionally allowed to authenticate

## Example

Let's create a struct `MyStruct` with a method `Authenticate()` which will return either an error or an azure token:

```go
import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/quara-dev/beyond/pkg/azutils"
)

type MyStruct struct {
    // Keep a reference to some credential config in the struct:
	AzCredential *azutils.CredentialConfig `json:"credential,omitempty"`
}

// Authenticate is the method exposed by `MyStruct` to obtain an azure token
func (m *MyStruct) Authenticate() (azcore.TokenCredential, error) {
	// Make sure credential exist:
	if m.AzCredential == nil {
		m.AzCredential = new(azutils.CredentialConfig)
	}
	// Build credential:
    // Optionally use .ParseEnv() before calling .Build()
    // to update credential according to environment variables
	login, err := m.AzCredential().ParseEnv().Build()
	if err != nil {
		return nil, err
	}
	// Authenticate using credential:
	return login.Authenticate()
}
```

## Caddyfile support

`CredentialConfig` implements the `caddyfile.Unmarshaler` interface and can be unmarshalled from caddyfile dispenser. It makes it easy to write `UnmarshalCaddyfile` for your own structs.

Let's update the struct defined above to parse credential from Caddyfile, like this example below:
 
```
{
    credential {
        client_id test
        client_secret test
    }
}
```

To do that, we need to define the `UnmarshalCaddyfile(d *caddyfile.Dispenser) error` method:

```go
type MyStruct struct {
    // Name is a required field in this example
    Name string `json:"name"`
    // Keep a reference to some credential config in the struct:
	AzCredential *azutils.CredentialConfig `json:"credential,omitempty"`
}

// UnmarshalCaddyfile can be used to unmarshal Mystruct{} from a caddyfile dispenser
func (m *MyStruct) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    for nesting := d.Nesting(); d.NextBlock(nesting); {
        directive := d.Val()
        switch directive {
            // Parse Name when directive is "name"
            case "name":
                if !d.AllArgs(&m.Name) {
                    return d.Err("expected name")
                }
            // Parse CredentialConfig when directive is "credential"
            case "credential":
                // Make sure AzCredential field is not nil
                if m.AzCredential == nil {
                    m.AzCredential = new(azutils.CredentialConfig)
                }
                // Unmarshal credential
                if err := m.AzCredential.UnmarshalCaddyfile(d); err != nil {
                    return err
                }
            // Return an error for unknown directives
            default:
                return errors.New("unrecognized subdirective: %s", directive)
        }
    }
    return nil
}
```
