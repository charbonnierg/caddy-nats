package mongo

import (
	"fmt"
	"net/url"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	connectors "github.com/quara-dev/beyond/modules/nats/connectors"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ChangeStreamExporter{})
}

type ChangeStreamExporter struct {
	ctx      caddy.Context
	client   *mongo.Client
	database *mongo.Database
	logger   *zap.Logger

	Uri      string `json:"uri"`
	Database string `json:"database"`
}

func (ChangeStreamExporter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.exporters.mongodb_change_stream",
		New: func() caddy.Module { return new(ChangeStreamExporter) },
	}
}

// Provision the exporter
func (e *ChangeStreamExporter) Provision(ctx caddy.Context) error {
	e.ctx = ctx
	e.logger = ctx.Logger().Named("exporter.mongodb_change_stream")
	e.logger.Info("provisioning mongodb change stream exporter", zap.String("uri", e.Uri))
	return nil
}

// Connect to the database
func (e *ChangeStreamExporter) Connect(clients *resources.Clients) error {
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
func (e *ChangeStreamExporter) Close() error {
	if e.client != nil {
		e.logger.Info("Disconnecting mongodb client")
		return e.client.Disconnect(e.ctx)
	}
	return nil
}

// Write writes the change event to the database.
func (e *ChangeStreamExporter) Write(rawMsg connectors.Message) error {
	data, err := rawMsg.Payload()
	if err != nil {
		return err
	}
	msg, err := NewChangeStreamEvent(data)
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

func (e *ChangeStreamExporter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "uri":
			if err := parser.ParseString(d, &e.Uri); err != nil {
				return err
			}
		case "database":
			if err := parser.ParseString(d, &e.Database); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}

var (
	_ connectors.Exporter = (*ChangeStreamExporter)(nil)
)
