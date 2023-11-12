package caddynats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"go.uber.org/zap"
)

type Message interface {
	Subject(prefix string) (string, error)
	Payload() ([]byte, error)
	Headers() (map[string][]string, error)
}

func NewNatsMessage(msg *nats.Msg) *NatsMessage {
	return &NatsMessage{msg: msg}
}

type NatsMessage struct {
	msg *nats.Msg
}

func (m *NatsMessage) Subject(prefix string) (string, error) {
	if prefix == "" {
		return m.msg.Subject, nil
	}
	return fmt.Sprintf("%s.%s", prefix, m.msg.Subject), nil
}

func (m *NatsMessage) Payload() ([]byte, error) {
	return m.msg.Data, nil
}

func (m *NatsMessage) Headers() (map[string][]string, error) {
	return m.msg.Header, nil
}

func NewJetStreamMessage(msg jetstream.Msg) *JetStreamMessage {
	return &JetStreamMessage{msg: msg}
}

type JetStreamMessage struct {
	msg jetstream.Msg
}

func (m *JetStreamMessage) Subject(prefix string) (string, error) {
	if prefix == "" {
		return m.msg.Subject(), nil
	}
	return fmt.Sprintf("%s.%s", prefix, m.msg.Subject()), nil
}

func (m *JetStreamMessage) Payload() ([]byte, error) {
	return m.msg.Data(), nil
}

func (m *JetStreamMessage) Headers() (map[string][]string, error) {
	return m.msg.Headers(), nil
}

type Reader interface {
	caddy.Module
	Open(ctx caddy.Context, client *natsclient.NatsClient) error
	Close() error
	Read() (Message, func() error, error)
}

type Writer interface {
	caddy.Module
	Open(ctx caddy.Context, client *natsclient.NatsClient) error
	Close() error
	Write(Message) error
}

type Flow struct {
	ctx         caddy.Context
	current     caddy.Context
	cancel      context.CancelFunc
	done        chan struct{}
	logger      *zap.Logger
	account     *Account
	client      *natsclient.NatsClient
	server      *Server
	source      Reader
	destination Writer

	Source      json.RawMessage `json:"source,omitempty" caddy:"namespace=nats_server.readers inline_key=type"`
	Destination json.RawMessage `json:"destination,omitempty" caddy:"namespace=nats_server.writers inline_key=type"`
	Disabled    bool            `json:"disabled,omitempty"`
}

// Provision configures the receiver and the exporter
// and returns an error if any.
func (c *Flow) Provision(server *Server, account *Account) error {
	c.ctx = server.ctx
	c.logger = c.ctx.Logger().Named("flow")
	sunm, err := c.ctx.LoadModule(c, "Source")
	if err != nil {
		return err
	}
	source, ok := sunm.(Reader)
	if !ok {
		return errors.New("source is not a Receiver")
	}
	c.source = source
	dunm, err := c.ctx.LoadModule(c, "Destination")
	if err != nil {
		return err
	}
	destination, ok := dunm.(Writer)
	if !ok {
		return errors.New("destination is not an Exporter")
	}
	c.destination = destination
	return nil
}

func (c *Flow) tick() (bool, error) {
	// Read a message from source
	msg, ack, err := c.source.Read()
	if err != nil {
		// If error is EOF, return false to stop
		if err.Error() == "EOF" {
			return false, nil
		}
		// Else return error and true to keep going
		return true, err
	}
	// If there is no message, return false to stop
	if msg == nil {
		return false, nil
	}
	// Write message to destination
	if err := c.destination.Write(msg); err != nil {
		// Return error and true to keep going
		return true, err
	}
	// Acknowledge message
	if err := ack(); err != nil {
		// Return error and true to keep going
		return true, err
	}
	// Return true to keep going
	return true, nil
}

// Run starts the flow and blocks until the context is cancelled.
func (c *Flow) run(ctx caddy.Context, done chan struct{}, client *natsclient.NatsClient) error {
	defer func() {
		if err := client.Close(); err != nil {
			c.logger.Error("error closing nats client", zap.Error(err), zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
		}
		close(done)
	}()
	c.logger.Info("starting data flow", zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
	if err := client.Connect(); err != nil {
		c.logger.Error("error connecting nats client", zap.Error(err), zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
		return err
	}
	// Try connecting to source forever until it succeeds
	for {
		if err := c.source.Open(ctx, client); err != nil {
			c.logger.Error("error connecting source", zap.Error(err), zap.String("source", c.source.CaddyModule().String()), zap.String("retry_in", "2s"))
			// TODO: add exponential backoff
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second * 2):
				continue
			}
		}
		break
	}
	// Try connecting to destination forever until it succeeds
	for {
		if err := c.destination.Open(ctx, client); err != nil {
			c.logger.Error("error connecting destination", zap.Error(err), zap.String("destination", c.destination.CaddyModule().String()), zap.String("retry_in", "2s"))
			// TODO: add exponential backoff
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second * 2):
				continue
			}
		}
		break
	}
	// Start ticking
	for {
		select {
		case <-ctx.Done():
			c.logger.Warn("data flow is stopped", zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
			return nil
		default:
			c.logger.Debug("data flow tick", zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
			keepgoing, err := c.tick()
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				c.logger.Error(err.Error(), zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
			}
			if !keepgoing {
				return nil
			}
		}
	}
}

func (c *Flow) Start() error {
	if c.Disabled {
		return nil
	}
	ctx, cancel := caddy.NewContext(c.ctx)
	client := &natsclient.NatsClient{Internal: true}
	if err := c.server.provisionInternalClientConnection(c.account.Name, client); err != nil {
		return err
	}
	c.done = make(chan struct{})
	c.current = ctx
	c.cancel = cancel
	go c.run(ctx, c.done, c.client)
	return nil
}

func (c *Flow) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.done != nil {
		<-c.done
	}
	return nil
}

func (flow *Flow) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "from", "source":
			var module string
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "nats_server.readers."+module)
			if err != nil {
				return err
			}
			flow.Source = caddyconfig.JSONModuleObject(unm, "type", module, nil)
		case "to", "destination":
			var module string
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "nats_server.writers."+module)
			if err != nil {
				return err
			}
			flow.Destination = caddyconfig.JSONModuleObject(unm, "type", module, nil)
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
