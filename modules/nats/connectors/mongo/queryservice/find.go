package mongoservice

import (
	"github.com/nats-io/nats.go/micro"
	"go.uber.org/zap"
)

// FindEndpoint is the handler for the "find" endpoint.
type FindEndpoint struct {
	logger *zap.Logger
}

// Handle handles the request.
func (s *FindEndpoint) Handle(req micro.Request) {
	s.logger.Info("handling request")
}
