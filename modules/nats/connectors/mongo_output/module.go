package mongo_output

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(MongoOutputConnector{})
}

// MongoOutputConnector is a Caddy module that serves as a connector
// to MongoDB. It writes data coming from a stream to a MongoDB database.
type MongoOutputConnector struct {
	logger         *zap.Logger
	URI            string                `json:"uri"`
	Database       string                `json:"database"`
	Client         *natsutils.Connection `json:"client"`
	StreamName     string                `json:"stream_name"`
	ConsumerName   string                `json:"consumer_name"`
	StartSequence  int64                 `json:"start_sequence"`
	StartTimestamp int64                 `json:"start_timestamp"`
}

// CaddyModule returns the Caddy module information.
func (MongoOutputConnector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "connectors.output.mongo",
		New: func() caddy.Module { return new(MongoOutputConnector) },
	}
}

func (c *MongoOutputConnector) Provision(app nats.App) error {
	c.logger = app.Logger().Named("connectors.output.mongo")

	return nil
}

func (c *MongoOutputConnector) Start() error {
	return nil
}

func (c *MongoOutputConnector) Stop() error {
	return nil
}

var (
	_ nats.Connector = (*MongoOutputConnector)(nil)
)
