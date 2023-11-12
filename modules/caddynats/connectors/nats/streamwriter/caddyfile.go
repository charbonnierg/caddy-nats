package streamexporter

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
)

func (s *StreamExporter) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if s.Stream == nil {
		s.Stream = new(natsclient.Stream)
	}
	if err := s.Stream.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	return nil
}
