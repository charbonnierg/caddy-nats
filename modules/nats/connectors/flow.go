package connectors

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
	"go.uber.org/zap"
)

// ReceiverService wraps any Receiver implementation and writes
// messages received by the receiver to a JetStream stream.
type Flow struct {
	ctx         caddy.Context
	logger      *zap.Logger
	source      Receiver
	destination Exporter

	Source      json.RawMessage `json:"source,omitempty" caddy:"namespace=nats.receivers inline_key=type"`
	Destination json.RawMessage `json:"destination,omitempty" caddy:"namespace=nats.exporters inline_key=type"`
}

// Provision configures the receiver and the exporter
// and returns an error if any.
func (c *Flow) Provision(ctx caddy.Context) error {
	c.ctx = ctx
	c.logger = ctx.Logger().Named("receiver")
	sunm, err := ctx.LoadModule(c, "Source")
	if err != nil {
		return err
	}
	source, ok := sunm.(Receiver)
	if !ok {
		return errors.New("source is not a Receiver")
	}
	c.source = source
	if err := c.source.Provision(ctx); err != nil {
		return err
	}
	dunm, err := ctx.LoadModule(c, "Destination")
	if err != nil {
		return err
	}
	destination, ok := dunm.(Exporter)
	if !ok {
		return errors.New("destination is not an Exporter")
	}
	c.destination = destination
	if err := c.destination.Provision(ctx); err != nil {
		return err
	}
	return nil
}

// Run starts the flow and blocks until the context is cancelled.
func (c *Flow) Run(clients *resources.Clients) error {
	next := make(chan bool, 1)
	var nextMsg Message
	var ack func() error
	c.logger.Info("running flow")
	if err := c.source.Connect(clients); err != nil {
		return err
	}
	if err := c.destination.Connect(clients); err != nil {
		return err
	}
	// Kick off the receiver
	next <- true
	for {
		select {
		// Exit if context is cancelled
		case <-c.ctx.Done():
			c.logger.Warn("context cancelled")
			return nil
		// Handle next boolean value
		case ok := <-next:
			switch ok {
			// If true, either write, acknowledge or read a message
			case true:
				switch {
				case nextMsg != nil:
					if err := c.destination.Write(nextMsg); err != nil {
						c.logger.Warn("error writing message", zap.Error(err))
					} else {
						nextMsg = nil
					}
					next <- true
				case ack != nil:
					if err := ack(); err != nil {
						c.logger.Warn("error acknowledging message", zap.Error(err))
					} else {
						ack = nil
					}
					next <- true
				default:
					var err error
					nextMsg, ack, err = c.source.Read()
					if err != nil {
						c.logger.Warn("error reading message", zap.Error(err))
						time.Sleep(time.Duration(1) * time.Second)
					}
					next <- true
				}
			// If false, stop the receiver
			case false:
				c.logger.Warn("stopping receiver")
				return nil
			}
		}
	}
}
