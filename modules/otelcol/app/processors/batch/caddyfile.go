package batch

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func (r *BatchProcessor) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
