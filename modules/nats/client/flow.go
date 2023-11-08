package client

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/caddyserver/caddy/v2"
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
	c.logger = ctx.Logger().Named("flow")
	sunm, err := ctx.LoadModule(c, "Source")
	if err != nil {
		return err
	}
	source, ok := sunm.(Receiver)
	if !ok {
		return errors.New("source is not a Receiver")
	}
	c.source = source
	if err := c.source.Provision(c.ctx); err != nil {
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
	if err := c.destination.Provision(c.ctx); err != nil {
		return err
	}
	return nil
}

func (c *Flow) tick() (bool, error) {
	msg, ack, err := c.source.Read()
	if err != nil {
		if err.Error() == "EOF" {
			return false, nil
		}
		return true, err
	}
	if msg == nil {
		return false, nil
	}
	if err := c.destination.Write(msg); err != nil {
		return true, err
	}
	if err := ack(); err != nil {
		return true, err
	}
	return true, nil
}

// Run starts the flow and blocks until the context is cancelled.
func (c *Flow) Run(clients *Clients) error {
	c.logger.Info("running flow")
	for {
		if err := c.source.Connect(clients); err != nil {
			c.logger.Error("error connecting source", zap.Error(err), zap.String("source", c.source.CaddyModule().String()))
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		break
	}
	for {
		if err := c.destination.Connect(clients); err != nil {
			c.logger.Error("error connecting destination", zap.Error(err), zap.String("destination", c.destination.CaddyModule().String()))
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		break
	}
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("context cancelled", zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
			return nil
		default:
			c.logger.Info("ticking", zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
			keepgoing, err := c.tick()
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				c.logger.Error("error ticking", zap.Error(err), zap.String("source", c.source.CaddyModule().String()), zap.String("destination", c.destination.CaddyModule().String()))
				select {
				case <-c.ctx.Done():
					return nil
				case <-time.After(time.Second):
					continue
				}
			}
			if !keepgoing {
				return nil
			}
		}
	}
}
