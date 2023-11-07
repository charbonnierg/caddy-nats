package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/connectors/resources"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func ParseStream(d *caddyfile.Dispenser, stream *resources.Stream) error {
	if stream.StreamConfig == nil {
		stream.StreamConfig = &nats.StreamConfig{}
	}
	if err := parser.ParseString(d, &stream.Name); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "prefix":
			if err := parser.ParseString(d, &stream.Prefix); err != nil {
				return err
			}
		case "subjects":
			if err := parser.ParseStringArray(d, &stream.Subjects); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
