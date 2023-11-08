package client

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats.go"
	natsapp "github.com/quara-dev/beyond/modules/nats"
	"go.uber.org/zap"
)

type Connections []*Connection

func (c Connections) Provision(app natsapp.App) error {
	for _, conn := range c {
		if err := conn.Provision(app); err != nil {
			return err
		}
	}
	return nil
}

// Connection is a struct that contains various services that can be run
// on a NATS connection. It can be used to create streams, request-reply services,
// receiver services and exporter services.
type Connection struct {
	ctx      caddy.Context
	logger   *zap.Logger
	wg       *sync.WaitGroup
	conn     *Clients
	services []*ServiceDefinition

	NatsClient
	Account        string            `json:"account,omitempty"`
	Streams        []*Stream         `json:"streams,omitempty"`
	KeyValueStores []*KeyValueStore  `json:"key_value_stores,omitempty"`
	ObjectStores   []*ObjectStore    `json:"object_stores,omitempty"`
	DataFlows      []*Flow           `json:"flows,omitempty"`
	Services       []json.RawMessage `json:"services,omitempty" caddy:"namespace=connectors.services inline_key=type"`
}

// Provision sets up the connection. It is mainly used to acquire a logger and a context.
func (a *Connection) Provision(app natsapp.App) error {
	a.ctx = app.Context()
	a.wg = &sync.WaitGroup{}
	a.logger = app.Logger()
	a.logger.Info("provisioning connector")
	// Consider client as internal if no servers are specified
	if a.NatsClient.Servers == nil {
		a.NatsClient.Internal = true
	}
	a.NatsClient.Provider = app
	// Set default client name
	if a.NatsClient.Name == "" && a.Name != "" {
		a.NatsClient.Name = a.Name
	}
	if a.Account != "" {
		token, err := app.AddNewTokenBasedAuthPolicy(a.Account)
		if err != nil {
			return err
		}
		a.NatsClient.Token = token
	}
	// Validate client
	if err := a.NatsClient.Validate(); err != nil {
		return err
	}
	// Load services
	unm, err := a.ctx.LoadModule(a, "Services")
	if err != nil {
		return err
	}
	for _, service := range unm.([]interface{}) {
		service, ok := service.(ServiceProvider)
		if !ok {
			return errors.New("service is not a ServiceProvider")
		}
		if err := service.Provision(a.ctx); err != nil {
			return err
		}
		definition, err := service.Definition()
		if err != nil {
			return err
		}
		a.services = append(a.services, definition)
	}
	// Provision components
	for _, components := range a.DataFlows {
		if err := components.Provision(a.ctx); err != nil {
			return err
		}
	}
	return nil
}

// Connect connects to the NATS server and configures the resources owned by the connection.
// It returns an error if the connection fails or if a resource configuration fails.
func (a *Connection) Connect() error {
	a.logger.Info("connecting to NATS server")
	// First establish connection
	conn, err := a.NatsClient.Connect()
	if err != nil {
		a.logger.Error("error connecting to NATS", zap.Error(err))
		return err
	}
	a.conn = conn
	// Configure streams
	for _, stream := range a.Streams {
		a.logger.Info("configuring stream", zap.String("name", stream.StreamConfig.Name))
		if err := stream.Configure(a.ctx, conn); err != nil {
			a.logger.Error("error configuring stream", zap.Error(err))
			return err
		}
	}
	// Configure key-value stores
	for _, kvstore := range a.KeyValueStores {
		a.logger.Info("configuring key-value store", zap.String("name", kvstore.Bucket))
		if err := kvstore.Configure(a.ctx, conn); err != nil {
			a.logger.Error("error configuring key-value store", zap.Error(err))
			return err
		}
	}
	// Configure object stores
	for _, objstore := range a.ObjectStores {
		a.logger.Info("configuring object store", zap.String("name", objstore.Bucket))
		if err := objstore.Configure(a.ctx, conn); err != nil {
			a.logger.Error("error configuring object store", zap.Error(err))
			return err
		}
	}
	// Run request-reply services
	for _, definition := range a.services {
		a.logger.Info("starting service", zap.String("name", definition.Name))
		if _, err := definition.Start(conn.Core()); err != nil {
			return err
		}
	}
	// Run exporters and receivers
	for _, flow := range a.DataFlows {
		a.wg.Add(1)
		go func(flow *Flow) {
			defer a.wg.Done()
			if err := flow.Run(conn); err != nil {
				a.logger.Error("data flow stopped", zap.Error(err))
			}
		}(flow)
	}
	return nil
}

// Close closes the connection. It returns an error if the connection fails to close.
func (a *Connection) Close() error {
	a.wg.Wait()
	return nil
}

// Conn returns the NATS client for this connection.
// If connection is not established, it will try to connect first.
func (a *Connection) Conn() (*nats.Conn, error) {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			return nil, err
		}
	}
	return a.conn.Core(), nil
}

// JetStream returns the JetStream context for this connection.
// If connection is not established, it will try to connect first.
func (a *Connection) JetStream() (nats.JetStreamContext, error) {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			return nil, err
		}
	}
	return a.conn.JetStream(), nil
}

// // AddOrUpdateStream adds or updates a stream.
// func (a *Connection) AddOrUpdateStream(stream *Stream) error {
// 	return stream.Configure(a.ctx, a.conn)
// }

// // AddOrUpdateConsumer adds or updates a consumer.
// func (a *Connection) AddOrUpdateConsumer(consumer *Consumer) error {
// 	return consumer.Configure(a.ctx, a.conn)
// }

// // AddOrUpdateKeyValueStore adds or updates a key-value store.
// func (a *Connection) AddOrUpdateKeyValueStore(kvstore *KeyValueStore) error {
// 	return kvstore.Configure(a.ctx, a.conn)
// }

// // AddOrUpdateObjectStore adds or updates an object store.
// func (a *Connection) AddOrUpdateObjectStore(objstore *ObjectStore) error {
// 	return objstore.Configure(a.ctx, a.conn)
// }

// // AddOrUpdateService adds or updates a service.
// func (a *Connection) AddOrUpdateService(service *ServiceDefinition) error {
// 	_, err := service.Start(a.conn.Core())
// 	return err
// }
