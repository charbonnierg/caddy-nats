package jetstream

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
	"github.com/quara-dev/beyond/pkg/natsutils"
)

func (s *JetStreamStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "name":
				if !d.AllArgs(&s.Name) {
					return d.ArgErr()
				}
			case "ttl":
				val, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				s.TTL = val
			case "client":
				s.Client = &natsutils.Client{}
				err := natscaddyfile.ParseClient(d, s.Client)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
