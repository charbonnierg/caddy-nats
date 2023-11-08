package consumerreceiver

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (r *StreamConsumerReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	if r.Consumer == nil {
		r.Consumer = new(client.Consumer)
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
