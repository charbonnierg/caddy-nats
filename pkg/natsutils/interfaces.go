package natsutils

import "github.com/nats-io/nats-server/v2/server"

// Keystore is an interface for a keystore that can be used to retrieve the auth signing key for an account
type Keystore interface {
	Get(account string) (string, error)
}

// InternalProvider is an interface for a provider that can be used to retrieve the internal NATS server
type InternalProvider interface {
	GetServer() (*server.Server, error)
}
