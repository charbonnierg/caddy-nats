package mongoservice

import (
	"github.com/nats-io/nats.go/micro"
	"go.uber.org/zap"
)

// FindHandler is the handler for the "find" endpoint.
type FindHandler struct {
	logger *zap.Logger
}

// Handle handles the request.
func (s *FindHandler) Handle(req micro.Request) {
	s.logger.Info("handling request")
}
