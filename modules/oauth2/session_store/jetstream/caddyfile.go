package jetstream

import (
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/oauth2/session_store/jetstream/internal"
)

func (s *JetStreamStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "ttl":
				val, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				s.TTL = val
			case "internal":
				makeClient(s)
				if !d.NextArg() {
					s.Client.Internal = true
				} else {
					val, err := parseBool(d)
					if err != nil {
						return err
					}
					s.Client.Internal = val
				}
			case "name":
				makeClient(s)
				if !d.AllArgs(&s.Client.Name) {
					return d.ArgErr()
				}
			case "servers":
				makeClient(s)
				if s.Client.Servers == nil {
					s.Client.Servers = []string{}
				}
				for d.NextArg() {
					if val := d.Val(); val != "" {
						s.Client.Servers = append(s.Client.Servers, val)
					}
				}
			case "username":
				makeClient(s)
				if !d.AllArgs(&s.Client.Username) {
					return d.ArgErr()
				}
			case "password":
				makeClient(s)
				if !d.AllArgs(&s.Client.Password) {
					return d.ArgErr()
				}
			case "token":
				makeClient(s)
				if !d.AllArgs(&s.Client.Token) {
					return d.ArgErr()
				}
			case "credentials":
				makeClient(s)
				if !d.AllArgs(&s.Client.Credentials) {
					return d.ArgErr()
				}
			case "seed":
				makeClient(s)
				if !d.AllArgs(&s.Client.Seed) {
					return d.ArgErr()
				}
			case "jwt":
				makeClient(s)
				if !d.AllArgs(&s.Client.Jwt) {
					return d.ArgErr()
				}
			case "jetstream_domain":
				makeClient(s)
				if !d.AllArgs(&s.Client.JSDomain) {
					return d.ArgErr()
				}
			case "jetstream_prefix":
				makeClient(s)
				if !d.AllArgs(&s.Client.JSPrefix) {
					return d.ArgErr()
				}
			case "inbox_prefix":
				makeClient(s)
				if !d.AllArgs(&s.Client.InboxPrefix) {
					return d.ArgErr()
				}
			case "no_randomize":
				makeClient(s)
				val, err := parseBool(d)
				if err != nil {
					return err
				}
				s.Client.NoRandomize = val
			case "ping_interval":
				makeClient(s)
				val, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return err
				}
				s.Client.PingInterval = val
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	return nil
}

func parseBool(d *caddyfile.Dispenser) (bool, error) {
	if !d.NextArg() {
		return true, nil
	}
	raw := d.Val()
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return false, d.Errf("invalid boolean value: %s", raw)
	}
	return val, nil
}

func makeClient(s *JetStreamStore) {
	if s.Client == nil {
		s.Client = &internal.Client{}
	}
}
