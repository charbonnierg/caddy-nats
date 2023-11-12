package consumerreader

import (
	"errors"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(StreamConsumerReader{})
}

type StreamConsumerReader struct {
	ctx    caddy.Context
	logger *zap.Logger
	con    jetstream.Consumer

	*natsclient.Consumer
}

func (StreamConsumerReader) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.readers.stream_consumer",
		New: func() caddy.Module { return new(StreamConsumerReader) },
	}
}

func (r *StreamConsumerReader) Read() (caddynats.Message, func() error, error) {
	for {
		batch, err := r.con.Fetch(1)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch messages: %s", err.Error())
		}
		select {
		case <-r.ctx.Done():
			return nil, nil, errors.New("EOF")
		case msg := <-batch.Messages():
			return caddynats.NewJetStreamMessage(msg), func() error { return msg.Ack() }, nil
		}
	}
}

func (r *StreamConsumerReader) Open(ctx caddy.Context, client *natsclient.NatsClient) error {
	r.logger.Info("provisioning NATS stream consumer reader", zap.String("name", r.Durable), zap.String("stream", r.Stream))
	r.ctx = ctx
	r.logger = ctx.Logger().Named("readers.stream_consumer")
	if err := client.ConfigureConsumer(ctx, r.Consumer); err != nil {
		return err
	}
	// Make sure stream consumer pull subscription is started
	js, err := client.JetStream()
	if err != nil {
		return err
	}
	con, err := js.Consumer(ctx, r.Stream, r.Durable)
	if err != nil {
		return err
	}
	r.con = con
	return nil
}

func (r *StreamConsumerReader) Close() error {
	defer func() {
		r.con = nil
	}()
	return nil
}

var (
	_ caddynats.Reader = (*StreamConsumerReader)(nil)
)
