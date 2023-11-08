package caddyfile

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func ParseServiceConnection(d *caddyfile.Dispenser, c *client.Connection) error {
	var service string
	if err := parser.ParseString(d, &service); err != nil {
		return err
	}
	unm, err := caddyfile.UnmarshalModule(d, "nats.services."+service)
	if err != nil {
		return err
	}
	s, ok := unm.(client.ServiceProvider)
	if !ok {
		return d.Errf("service '%s' invalid type", service)
	}
	c.Services = fnutils.DefaultIfEmpty(c.Services, []json.RawMessage{})
	c.Services = append(c.Services, caddyconfig.JSONModuleObject(s, "type", service, nil))
	return nil
}
