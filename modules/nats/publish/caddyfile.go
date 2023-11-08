package publish

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
)

func parseHandlerDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	p := NatsPublish{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

// UnmarshalCaddyfile parses the "nats_publish" directive from
// a Caddyfile dispenser.
// Syntax:
//
//	  nats_publish {
//		   connection <name> <type> {
//		      [options]
//		   }
//		   subject <subject>
//	  }
//
// Example:
//
//	  nats_publish {
//		   connection my-client tcp {
//		      urls tls://localhost:4222
//		   }
//		   subject some.subject
//	  }
func (p *NatsPublish) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	if d.Val() != "nats_publish" {
		return d.Errf("Expected 'nats_publish' directive, got '%s'", d.Val())
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "subject":
			sub := ""
			if !d.AllArgs(&sub) {
				return d.Err("invalid subject")
			}
			p.Subject = sub
		case "connection":
			if err := natscaddyfile.ParseConnection(d, &p.Connection); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
