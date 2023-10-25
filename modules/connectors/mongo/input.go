package mongo

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/damianiandrea/mongodb-nats-connector/pkg/connector"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/connectors"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(MongoInputConnector{})
}

type Collection struct {
	Database         string `json:"database"`
	Name             string `json:"collection"`
	PreAndPost       bool   `json:"pre_and_post"`
	TokensDBName     string `json:"tokens_db_name"`
	TokensCollName   string `json:"tokens_collection_name"`
	TokensCollCapped int64  `json:"tokens_collection_capped"`
	StreamName       string `json:"stream_name"`
}

// MongoInputConnector is a Caddy module that serves as a connector
// to MongoDB. It reads data from a MongoDB database and sends it
// to a stream.
type MongoInputConnector struct {
	conn        *connector.Connector
	NatsURL     string       `json:"nats_url"`
	URI         string       `json:"uri"`
	Collections []Collection `json:"collections"`
}

// CaddyModule returns the Caddy module information.
func (MongoInputConnector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "connectors.input.mongo",
		New: func() caddy.Module { return new(MongoInputConnector) },
	}
}

func (c *MongoInputConnector) Start() error {
	go c.conn.Run()
	return nil
}

func (c *MongoInputConnector) Stop() error {
	return nil
}

func (c *MongoInputConnector) Provision(app connectors.ConnectorsApp) error {
	ctx := app.Context()
	logger := ctx.Logger()
	opts := []connector.Option{
		connector.WithLogger(logger),
		connector.WithMongoUri(c.URI),
		connector.WithMongoOptions(),
		connector.WithNatsUrl(c.NatsURL),
		connector.WithNatsOptions(
			nats.MaxReconnects(-1),
			nats.RetryOnFailedConnect(true),
		), // your NATS options
		connector.WithContext(context.TODO()),
		connector.WithServerAddr(":10900"),
	}
	for _, coll := range c.Collections {
		colOpts := []connector.CollectionOption{}
		if coll.PreAndPost {
			colOpts = append(colOpts, connector.WithChangeStreamPreAndPostImages())
		}
		if coll.TokensDBName != "" {
			colOpts = append(colOpts, connector.WithTokensDbName(coll.TokensDBName))
		}
		if coll.TokensCollName != "" {
			colOpts = append(colOpts, connector.WithTokensCollName(coll.TokensCollName))
		}
		if coll.TokensCollCapped != 0 {
			colOpts = append(colOpts, connector.WithTokensCollCapped(coll.TokensCollCapped))
		}
		if coll.StreamName != "" {
			colOpts = append(colOpts, connector.WithStreamName(coll.StreamName))
		}
		opts = append(opts, connector.WithCollection(coll.Database, coll.Name, colOpts...))
	}
	logger.Info("Starting mongo connector")
	conn, err := connector.New(opts...)
	if err != nil {
		logger.Error("Error starting mongo connector", zap.Error(err))
		return err
	}
	c.conn = conn
	logger.Info("Mongo connector started")
	return nil
}

var (
	_ connectors.InputConnector = (*MongoInputConnector)(nil)
)
