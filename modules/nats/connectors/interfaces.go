package connectors

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
)

// Message is the entity that will be sent to the exporter or received from the receiver
type Message interface {
	Subject(prefix string) (string, error)
	Payload() ([]byte, error)
	Headers() (map[string][]string, error)
}

// Exporter is the interface that must be implemented by connectors writing to a
// third party service
type Exporter interface {
	caddy.Module
	Provision(ctx caddy.Context) error
	Connect(clients *resources.Clients) error
	Close() error
	Write(Message) error
}

// Receiver is the interface that must be implemented by connectors reading from a
// third party service
type Receiver interface {
	caddy.Module
	Provision(ctx caddy.Context) error
	Connect(clients *resources.Clients) error
	Close() error
	Read() (Message, func() error, error)
}

// ServiceProvider is the interface that must be implemented by connectors providing
// a third party service
type ServiceProvider interface {
	caddy.Module
	Provision(ctx caddy.Context) error
	Definition() (*resources.ServiceDefinition, error)
}
