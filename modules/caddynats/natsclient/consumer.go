package natsclient

import (
	"context"
	"errors"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

// Consumer is a JetStream consumer. Consumers are used to fetch messages from a stream
// and track the delivery state of those messages.
type Consumer struct {
	Stream string `json:"stream"`
	*jetstream.ConsumerConfig
}

// Configure creates or updates the consumer.
func (c *Consumer) Configure(ctx context.Context, js jetstream.JetStream) error {
	if c.ConsumerConfig == nil {
		return errors.New("consumer config is nil")
	}
	_, err := js.CreateOrUpdateConsumer(ctx, c.Stream, *c.ConsumerConfig)
	if err != nil {
		return err
	}
	return nil
}

func (c *Consumer) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if c.ConsumerConfig == nil {
		c.ConsumerConfig = &jetstream.ConsumerConfig{}
	}
	d.Next() // consume "consumer"
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "stream":
			if err := parser.ParseString(d, &c.Stream); err != nil {
				return err
			}
		case "description":
			if err := parser.ParseString(d, &c.Description); err != nil {
				return err
			}
		case "name", "durable", "durable_name", "consumer_name":
			if err := parser.ParseString(d, &c.Durable); err != nil {
				return err
			}
		case "deliver_policy":
			if err := parseDeliverPolicy(d, &c.DeliverPolicy); err != nil {
				return err
			}
		case "opt_start_seq":
			if err := parser.ParseUint64(d, &c.OptStartSeq); err != nil {
				return err
			}
		case "opt_start_time":
			if err := parser.ParseTime(d, c.OptStartTime); err != nil {
				return err
			}
		case "ack_policy":
			if err := parseAckPolicy(d, &c.AckPolicy); err != nil {
				return err
			}
		case "ack_wait":
			if err := parser.ParseDuration(d, &c.AckWait); err != nil {
				return err
			}
		case "max_deliver":
			if err := parser.ParseInt(d, &c.MaxDeliver); err != nil {
				return err
			}
		case "back_off":
			if c.BackOff == nil {
				c.BackOff = []time.Duration{}
			}
			for d.CountRemainingArgs() > 0 {
				var backOff time.Duration
				if err := parser.ParseDuration(d, &backOff); err != nil {
					return err
				}
				c.BackOff = append(c.BackOff, backOff)
			}
		case "filter_subject":
			if err := parser.ParseString(d, &c.FilterSubject); err != nil {
				return err
			}
		case "filter_subjects":
			if err := parser.ParseStringArray(d, &c.FilterSubjects); err != nil {
				return err
			}
		case "replay_policy":
			if err := parseReplayPolicy(d, &c.ReplayPolicy); err != nil {
				return err
			}
		case "rate_limit":
			if err := parser.ParseUint64(d, &c.RateLimit); err != nil {
				return err
			}
		case "sample_freq", "sample_frequency":
			if err := parser.ParseString(d, &c.SampleFrequency); err != nil {
				return err
			}
		case "max_waiting":
			if err := parser.ParseInt(d, &c.MaxWaiting); err != nil {
				return err
			}
		case "max_ack_pending":
			if err := parser.ParseInt(d, &c.MaxAckPending); err != nil {
				return err
			}
		case "headers_only":
			if err := parser.ParseBool(d, &c.HeadersOnly); err != nil {
				return err
			}
		case "max_batch":
			if err := parser.ParseInt(d, &c.MaxRequestBatch); err != nil {
				return err
			}
		case "max_expires":
			if err := parser.ParseDuration(d, &c.MaxRequestExpires); err != nil {
				return err
			}
		case "max_bytes":
			if err := parser.ParseInt(d, &c.MaxRequestMaxBytes); err != nil {
				return err
			}
		case "inactive_threshold":
			if err := parser.ParseDuration(d, &c.InactiveThreshold); err != nil {
				return err
			}
		case "num_replicas":
			if err := parser.ParseInt(d, &c.Replicas); err != nil {
				return err
			}
		case "mem_storage":
			if err := parser.ParseBool(d, &c.MemoryStorage); err != nil {
				return err
			}
		case "metadata":
			if err := parser.ParseStringMap(d, &c.Metadata); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized consumer config key '%s'", d.Val())
		}
	}
	return nil
}
