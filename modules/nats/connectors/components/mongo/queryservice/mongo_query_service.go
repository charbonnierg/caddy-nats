package mongoservice

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/nats/connectors"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(MongoQueryService{})
}

func (MongoQueryService) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.services.mongodb_query",
		New: func() caddy.Module { return new(MongoQueryService) },
	}
}

// MongoQueryService is a service that allows to query a MongoDB database.
type MongoQueryService struct {
	logger *zap.Logger

	Uri        string `json:"uri,omitempty"`
	QueueGroup string `json:"queue_group,omitempty"`
}

// Provision sets up the service. It is required by the connectors.ServiceProvider interface.
func (s *MongoQueryService) Provision(ctx caddy.Context) error {
	s.logger = ctx.Logger().Named("service.mongodb_query")
	return nil
}

// Definition returns the service definition. It is required by the connectors.ServiceProvider interface.
func (s *MongoQueryService) Definition() (*resources.ServiceDefinition, error) {
	def := &resources.ServiceDefinition{
		Name:        "MongoDB-Query-Service",
		Version:     "0.0.1",
		Description: `This service allows to query a MongoDB database.`,
		Endpoints: []*resources.EndpointDefinition{
			// Add endpoints here
			{
				Name:       "query",
				Subject:    "api.mongo.find",
				Handler:    &FindHandler{logger: s.logger},
				QueueGroup: s.QueueGroup,
			},
		},
	}
	return def, nil
}

// Interface guards
var (
	_ connectors.ServiceProvider = (*MongoQueryService)(nil)
)
