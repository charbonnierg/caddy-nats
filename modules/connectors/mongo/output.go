package mongo

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/connectors"
)

func init() {
	caddy.RegisterModule(MongoOutputConnector{})
}

// MongoOutputConnector is a Caddy module that serves as a connector
// to MongoDB. It writes data coming from a stream to a MongoDB database.
type MongoOutputConnector struct{}

// CaddyModule returns the Caddy module information.
func (MongoOutputConnector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "connectors.output.mongo",
		New: func() caddy.Module { return new(MongoOutputConnector) },
	}
}

func (c *MongoOutputConnector) Provision(app connectors.ConnectorsApp) error {
	return nil
}

func (c *MongoOutputConnector) Start() error {
	return nil
}

func (c *MongoOutputConnector) Stop() error {
	return nil
}

var (
	_ connectors.OutputConnector = (*MongoOutputConnector)(nil)
)
