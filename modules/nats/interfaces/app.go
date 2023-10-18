package interfaces

import (
	"github.com/nats-io/nats-server/v2/server"
)

type NatsApp interface {
	ServerProvider
}

type ServerProvider interface {
	GetServer() (*server.Server, error)
}
