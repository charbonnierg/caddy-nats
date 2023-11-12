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
	c.account = account
	c.server = server
	ctx, cancel := caddy.NewContext(server.ctx)
	c.client = &natsclient.NatsClient{Internal: true}
	if _, err := c.server.createInternalClientForAccount(c.account, c.client); err != nil {
		return err
	}
	c.ctx = ctx
	c.cancel = cancel
	c.done = make(chan struct{})
	c.logger = server.logger.Named("flow")
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
func (c *Flow) run() error {
	defer func() {
		if err := c.client.Close(); err != nil {
			c.logger.Error("error closing nats client", zap.Error(err), zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
		}
		close(c.done)
	}()
	c.logger.Info("starting data flow", zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
	if err := c.client.Connect(); err != nil {
		c.logger.Error("error connecting nats client", zap.Error(err), zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
		return err
	}
	// Try connecting to source forever until it succeeds
	for {
		if err := c.source.Open(c.ctx, c.client); err != nil {
			c.logger.Error("error connecting source", zap.Error(err), zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("retry_in", "2s"))
			// TODO: add exponential backoff
			select {
			case <-c.ctx.Done():
				return nil
			case <-time.After(time.Second * 2):
				continue
			}
		}
		break
	}
	// Try connecting to destination forever until it succeeds
	for {
		if err := c.destination.Open(c.ctx, c.client); err != nil {
			c.logger.Error("error connecting destination", zap.Error(err), zap.String("destination", c.destination.CaddyModule().ID.Name()), zap.String("retry_in", "2s"))
			// TODO: add exponential backoff
			select {
			case <-c.ctx.Done():
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
		case <-c.ctx.Done():
			c.logger.Warn("data flow is stopped", zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
			return nil
		default:
			c.logger.Debug("waiting for next message", zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
			keepgoing, err := c.tick()
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				c.logger.Error(err.Error(), zap.String("source", c.source.CaddyModule().ID.Name()), zap.String("destination", c.destination.CaddyModule().ID.Name()))
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

	go c.run()
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
