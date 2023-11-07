package nats

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/connectors"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(StreamConsumerReceiver{})
}

type StreamConsumerReceiver struct {
	ctx    caddy.Context
	logger *zap.Logger
	sub    *nats.Subscription

	*resources.Consumer
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

func (r *StreamConsumerReceiver) Read() (connectors.Message, func() error, error) {
	for {
		msgs, err := r.sub.Fetch(1, nats.MaxWait(time.Duration(1)*time.Second))
		if err != nil {
			if err == nats.ErrTimeout {
				// No messages available, continue
				continue
			}
			// Will be retried
			return nil, nil, err
		}
		message := msgs[0]
		return &natsMessage{msg: message}, func() error { return message.Ack() }, nil
	}
}

func (r *StreamConsumerReceiver) Connect(clients *resources.Clients) error {
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
		nats.BindStream(r.Stream),
		nats.Context(r.ctx),
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

func (r *StreamConsumerReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	if r.Consumer == nil {
		r.Consumer = new(resources.Consumer)
	}
	if r.Consumer.ConsumerConfig == nil {
		r.Consumer.ConsumerConfig = new(nats.ConsumerConfig)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "stream":
			if err := parser.ParseString(d, &r.Stream); err != nil {
				return err
			}
		case "durable", "consumer_name":
			if err := parser.ParseString(d, &r.Durable); err != nil {
				return err
			}
		case "filter_subject":
			if err := parser.ParseString(d, &r.FilterSubject); err != nil {
				return err
			}
		case "filter_subjects":
			if err := parser.ParseStringArray(d, &r.FilterSubjects); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
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
	_ connectors.Receiver = (*StreamConsumerReceiver)(nil)
)
