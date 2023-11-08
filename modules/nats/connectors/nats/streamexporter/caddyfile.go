package streamexporter

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (s *StreamExporter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	parser.ExpectString(d)
	if s.Stream == nil {
		s.Stream = new(client.Stream)
	}
	if s.Stream.StreamConfig == nil {
		s.Stream.StreamConfig = new(nats.StreamConfig)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "name":
			if err := parser.ParseString(d, &s.Name); err != nil {
				return err
			}
		case "prefix":
			if err := parser.ParseString(d, &s.Prefix); err != nil {
				return err
			}
		case "subjects":
			if err := parser.ParseStringArray(d, &s.StreamConfig.Subjects); err != nil {
				return err
			}
		}
	}
	return nil
}
