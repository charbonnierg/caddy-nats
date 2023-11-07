package natsapp

import (
	"encoding/json"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
)

type Config struct {
	AuthServiceRaw *auth.AuthServiceConfig `json:"auth_service,omitempty"`
	ServerRaw      *embedded.Options       `json:"server,omitempty"`
	ConnectorsRaw  []json.RawMessage       `json:"connectors,omitempty" caddy:"namespace=nats.connectors inline_key=module"`
	ReadyTimeout   time.Duration           `json:"ready_timeout,omitempty"`
}

func (a *Config) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// Make sure auth service exists in case auth policy are defined
	// within account blocks.
	if err := parser.ExpectString(d, parser.Match("nats", "broker")); err != nil {
		return err
	}
	a.AuthServiceRaw = fnutils.DefaultIfNil(a.AuthServiceRaw, &auth.AuthServiceConfig{})
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "debug":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			if err := parser.ParseBool(d, &a.ServerRaw.Debug); err != nil {
				return err
			}
		case "default_auth_callout":
			var name string
			if err := parser.ParseString(d, &name); err != nil {
				return err
			}
			mod, err := caddyfile.UnmarshalModule(d, "nats.auth_callout."+name)
			if err != nil {
				return d.Errf("failed to unmarshal module '%s': %v", name, err)
			}
			a.AuthServiceRaw.DefaultHandlerRaw = caddyconfig.JSONModuleObject(mod, "module", name, nil)
		case "auth_service":
			a.AuthServiceRaw = fnutils.DefaultIfNil(a.AuthServiceRaw, &auth.AuthServiceConfig{})
			if err := natscaddyfile.ParseAuthServiceConfig(d, a.AuthServiceRaw); err != nil {
				return err
			}
		case "accounts":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Accounts = fnutils.DefaultIfEmpty(a.ServerRaw.Accounts, []*embedded.Account{})
			if err := natscaddyfile.ParseAccounts(d, a.AuthServiceRaw, &a.ServerRaw.Accounts); err != nil {
				return err
			}
		case "account":
			acc := embedded.Account{}
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Accounts = fnutils.DefaultIfEmpty(a.ServerRaw.Accounts, []*embedded.Account{})
			if err := parser.ParseString(d, &acc.Name); err != nil {
				return err
			}
			if err := natscaddyfile.ParseAccount(d, a.AuthServiceRaw, &acc); err != nil {
				return err
			}
			a.ServerRaw.Accounts = append(a.ServerRaw.Accounts, &acc)
		case "jetstream":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.JetStream = fnutils.DefaultIfNil(a.ServerRaw.JetStream, &embedded.JetStream{})
			if err := natscaddyfile.ParseJetStream(d, a.ServerRaw.JetStream); err != nil {
				return err
			}
		case "nats_server", "server":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			if err := natscaddyfile.ParseOptions(d, a.ServerRaw); err != nil {
				return err
			}
		case "mqtt_server":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Mqtt = fnutils.DefaultIfNil(a.ServerRaw.Mqtt, &embedded.MQTT{})
			if err := natscaddyfile.ParseMqtt(d, a.ServerRaw.Mqtt); err != nil {
				return err
			}
		case "websocket_server":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Websocket = fnutils.DefaultIfNil(a.ServerRaw.Websocket, &embedded.Websocket{})
			if err := natscaddyfile.ParseWebsocket(d, a.ServerRaw.Websocket); err != nil {
				return err
			}
		case "leafnode", "leafnode_server":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Leafnode = fnutils.DefaultIfNil(a.ServerRaw.Leafnode, &embedded.Leafnode{})
			if err := natscaddyfile.ParseLeafnodes(d, a.ServerRaw.Leafnode); err != nil {
				return err
			}
		case "metrics":
			a.ServerRaw = fnutils.DefaultIfNil(a.ServerRaw, embedded.NewOptions())
			a.ServerRaw.Metrics = fnutils.DefaultIfNil(a.ServerRaw.Metrics, &embedded.Metrics{})
			if err := natscaddyfile.ParseMetrics(d, a.ServerRaw.Metrics); err != nil {
				return err
			}
		case "ready_timeout":
			if err := parser.ParseDuration(d, &a.ReadyTimeout); err != nil {
				return err
			}
		case "connector":
			var module string
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			mod, err := caddyfile.UnmarshalModule(d, "nats.connectors."+module)
			if err != nil {
				return d.Errf("failed to unmarshal module '%s': %v", module, err)
			}
			if a.ConnectorsRaw == nil {
				a.ConnectorsRaw = []json.RawMessage{}
			}
			a.ConnectorsRaw = append(a.ConnectorsRaw, caddyconfig.JSONModuleObject(mod, "module", module, nil))
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	// Remove empty auth service
	if a.AuthServiceRaw.Zero() {
		a.AuthServiceRaw = nil
	}
	return nil
}
