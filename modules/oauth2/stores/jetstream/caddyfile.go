package jetstream

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

func (s *JetStreamStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "name":
				if err := parser.ParseString(d, &s.Name); err != nil {
					return err
				}
			case "ttl":
				if err := parser.ParseDuration(d, &s.TTL); err != nil {
					return err
				}
			case "client":
				s.Client = &natsutils.Client{}
				if err := natscaddyfile.ParseClient(d, s.Client); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
