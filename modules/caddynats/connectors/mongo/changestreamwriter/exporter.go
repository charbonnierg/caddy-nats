package changestreamewriter

import (
	"fmt"
	"net/url"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/connectors/mongo/common"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ChangeStreamWriter{})
}

type ChangeStreamWriter struct {
	ctx      caddy.Context
	client   *mongo.Client
	database *mongo.Database
	logger   *zap.Logger

	Uri      string `json:"uri"`
	Database string `json:"database"`
}

func (ChangeStreamWriter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.writers.mongodb_change_stream",
		New: func() caddy.Module { return new(ChangeStreamWriter) },
	}
}

// Connect to the database
func (e *ChangeStreamWriter) Open(ctx caddy.Context, clients *natsclient.NatsClient) error {
	e.ctx = ctx
	e.logger = ctx.Logger().Named("exporter.mongodb_change_stream")
	e.logger.Info("provisioning mongodb change stream exporter", zap.String("uri", e.Uri))
	parsedUri, err := url.Parse(e.Uri)
	if err != nil {
		return fmt.Errorf("invalid mongodb uri: %v", err)
	}
	client, err := mongo.Connect(e.ctx, options.Client().ApplyURI(e.Uri))
	if err != nil {
		return fmt.Errorf("could not connect to mongodb: %v", err)
	}
	e.client = client
	e.logger.Info("connecting to mongodb", zap.String("uri", parsedUri.Redacted()))
	// Set database and collection
	e.database = e.client.Database(e.Database)
	return nil
}

// Close the connection
func (e *ChangeStreamWriter) Close() error {
	if e.client != nil {
		e.logger.Info("Disconnecting mongodb client")
		return e.client.Disconnect(e.ctx)
	}
	return nil
}

// Write writes the change event to the database.
func (e *ChangeStreamWriter) Write(rawMsg caddynats.Message) error {
	data, err := rawMsg.Payload()
	if err != nil {
		return err
	}
	msg, err := common.NewChangeStreamEvent(data)
	if err != nil {
		return err
	}
	col, err := msg.LookupCollection()
	if err != nil {
		return err
	}
	op, err := msg.LookupOperationType()
	if err != nil {
		return err
	}
	var writeModel mongo.WriteModel
	switch op {
	case "insert", "update", "delete":
		model, err := msg.WriteModel()
		if err != nil {
			return err
		}
		writeModel = model
	default:
		return fmt.Errorf("unsupported operation: %s", op)
	}
	collection := e.database.Collection(col)
	e.logger.Info("writing to mongodb", zap.String("collection", col), zap.String("operation", op))
	_, err = collection.BulkWrite(e.ctx, []mongo.WriteModel{writeModel})
	if err != nil {
		return fmt.Errorf("failed to write to mongo database: %s", err.Error())
	}
	return nil
}

var (
	_ caddynats.Writer = (*ChangeStreamWriter)(nil)
)
