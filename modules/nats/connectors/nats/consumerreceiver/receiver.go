package consumerreceiver

import (
	"context"
	"errors"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(StreamConsumerReceiver{})
}

type StreamConsumerReceiver struct {
	ctx    caddy.Context
	logger *zap.Logger
	sub    *nats.Subscription

	*client.Consumer
}

func (StreamConsumerReceiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.receivers.stream_consumer",
		New: func() caddy.Module { return new(StreamConsumerReceiver) },
	}
}

func (r *StreamConsumerReceiver) Provision(ctx caddy.Context) error {
	r.ctx = ctx
	r.logger = ctx.Logger().Named("receivers.stream_consumer")
	r.logger.Info("provisioning NATS stream consumer receiver", zap.String("name", r.Durable), zap.String("stream", r.Stream))
	return nil
}

func (r *StreamConsumerReceiver) Read() (client.Message, func() error, error) {
	for {
		deadline, cancel := context.WithTimeout(r.ctx, time.Second)
		msgs, err := r.sub.Fetch(1, nats.Context(deadline))
		cancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				// No messages available, continue
				continue
			}
			if err == context.Canceled {
				r.logger.Info("stopping stream consumer receiver")
				return nil, nil, errors.New("EOF")
			}
			// Will be retried
			return nil, nil, err
		}
		for _, msg := range msgs {
			return &natsMessage{msg: msg}, func() error { return msg.Ack() }, nil
		}
	}
}

func (r *StreamConsumerReceiver) Connect(clients *client.Clients) error {
	if err := r.Consumer.Configure(r.ctx, clients); err != nil {
		return err
	}
	// Make sure stream consumer pull subscription is started
	js := clients.JetStream()
	var subjects []string
	if r.FilterSubject != "" {
		subjects = []string{r.FilterSubject}
	} else {
		subjects = r.FilterSubjects
	}
	sub, err := js.PullSubscribe(
		"",
		r.Durable,
		nats.ConsumerFilterSubjects(subjects...),
		nats.Bind(r.Stream, r.Durable),
	)
	if err != nil {
		return err
	}
	r.sub = sub
	return nil
}

func (r *StreamConsumerReceiver) Close() error {
	defer func() {
		r.sub = nil
	}()
	if r.sub != nil {
		return r.sub.Unsubscribe()
	}
	return nil
}

type natsMessage struct {
	msg *nats.Msg
}

func (m *natsMessage) Payload() ([]byte, error) {
	return m.msg.Data, nil
}

func (m *natsMessage) Subject(prefix string) (string, error) {
	return m.msg.Subject, nil
}

func (m *natsMessage) Headers() (map[string][]string, error) {
	return m.msg.Header, nil
}

var (
	_ client.Receiver = (*StreamConsumerReceiver)(nil)
)
