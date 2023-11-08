package client

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
)

// Consumer is a JetStream consumer. Consumers are used to fetch messages from a stream
// and track the delivery state of those messages.
type Consumer struct {
	Stream string `json:"stream"`
	*nats.ConsumerConfig
}

// Configure creates or updates the consumer.
func (c *Consumer) Configure(ctx context.Context, clients *Clients) error {
	if c.ConsumerConfig == nil {
		return errors.New("consumer config is nil")
	}
	js := clients.JetStream()
	current, err := js.ConsumerInfo(c.Stream, c.ConsumerConfig.Durable, nats.Context(ctx))
	switch err {
	case nats.ErrConsumerNotFound:
		_, err = js.AddConsumer(c.Stream, c.ConsumerConfig, nats.Context(ctx))
		if err != nil {
			return err
		}
	case nil:
		if c.isApproximatelyEqualTo(&current.Config) {
			return nil
		}
		_, err = js.UpdateConsumer(c.Stream, c.ConsumerConfig, nats.Context(ctx))
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func (s *Consumer) isApproximatelyEqualTo(other *nats.ConsumerConfig) bool {
	if other == nil {
		return false
	}
	if s.Durable != other.Durable {
		return false
	}
	if s.DeliverSubject != other.DeliverSubject {
		return false
	}
	if s.DeliverPolicy != other.DeliverPolicy {
		return false
	}
	if s.AckPolicy != other.AckPolicy {
		return false
	}
	if s.AckWait != other.AckWait {
		return false
	}
	if s.MaxDeliver != other.MaxDeliver {
		return false
	}
	if s.FilterSubject != other.FilterSubject {
		return false
	}
	if s.ReplayPolicy != other.ReplayPolicy {
		return false
	}
	if s.SampleFrequency != other.SampleFrequency {
		return false
	}
	if s.Heartbeat != other.Heartbeat {
		return false
	}
	if s.MaxAckPending != other.MaxAckPending {
		return false
	}
	if s.MaxWaiting != other.MaxWaiting {
		return false
	}
	if s.FlowControl != other.FlowControl {
		return false
	}
	if s.RateLimit != other.RateLimit {
		return false
	}
	return true
}

var (
	_ resource = (*Consumer)(nil)
)
