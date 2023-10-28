package natsapp

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth"
	"github.com/quara-dev/beyond/modules/nats/embedded"
)

type Config struct {
	AuthServiceRaw *auth.AuthServiceConfig `json:"auth_service,omitempty"`
	ServerRaw      *embedded.Options       `json:"server,omitempty"`
	ReadyTimeout   time.Duration           `json:"ready_timeout,omitempty"`
}

func (a *Config) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "auth_service":
				if a.AuthServiceRaw == nil {
					a.AuthServiceRaw = new(auth.AuthServiceConfig)
				}
				if err := a.AuthServiceRaw.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			case "server":
				if a.ServerRaw == nil {
					a.ServerRaw = embedded.NewOptions()
				}
				if err := a.ServerRaw.UnmarshalCaddyfile(d); err != nil {
					return err
				}
			case "ready_timeout":
				if !d.NextArg() {
					return d.Err("expected ready timeout")
				}
				dur, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("failed to parse ready timeout: %v", err)
				}
				a.ReadyTimeout = dur
			default:
				return d.Errf("unknown directive '%s'", d.Val())
			}
		}
	}
	return nil
}
