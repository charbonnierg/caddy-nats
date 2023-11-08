package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func ParseFlow(d *caddyfile.Dispenser, flow *client.Flow) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "from":
			var module string
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "nats.receivers."+module)
			if err != nil {
				return err
			}
			flow.Source = caddyconfig.JSONModuleObject(unm, "type", module, nil)
		case "to":
			var module string
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "nats.exporters."+module)
			if err != nil {
				return err
			}
			flow.Destination = caddyconfig.JSONModuleObject(unm, "type", module, nil)
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
