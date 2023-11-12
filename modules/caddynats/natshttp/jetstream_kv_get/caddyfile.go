package jetstream_kv_get

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func parseHandlerDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	p := &JetStreamKeyValueGet{}
	err := p.UnmarshalCaddyfile(h.Dispenser)
	return p, err
}

// UnmarshalCaddyfile parses the "jetstream_publish" directive from
// a Caddyfile dispenser.
// Syntax:
//
//		  jetstream_kv_get {
//			   bucket <bucket>
//	           key <key>
//			   [account <account>]
//	           [client  {
//			   		[options]
//	           }]
//		  }
//
// Example:
//
//		  jetstream_kv_get {
//			   bucket some-bucket
//	           key some-key
//		  }
func (p *JetStreamKeyValueGet) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "bucket":
			bucket := ""
			if !d.AllArgs(&bucket) {
				return d.Err("invalid bucket")
			}
			p.Bucket = bucket
		case "key":
			key := ""
			if !d.AllArgs(&key) {
				return d.Err("invalid key")
			}
			p.Key = key
		case "account":
			account := ""
			if !d.AllArgs(&account) {
				return d.Err("invalid account")
			}
			p.Account = account
		case "client":
			p.Client = fnutils.DefaultIfNil(p.Client, &natsclient.NatsClient{})
			if err := p.Client.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
