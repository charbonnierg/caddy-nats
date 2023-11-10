// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsapp

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/modules/nats/auth"
	natscaddyfile "github.com/quara-dev/beyond/modules/nats/caddyfile"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	a := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, a)
		if err != nil {
			return nil, err
		}
	}
	err := a.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "nats",
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// Make sure auth service exists in case auth policy are defined
	// within account blocks.
	if err := parser.ExpectString(d, parser.Match("nats", "broker")); err != nil {
		return err
	}
	a.AuthService = fnutils.DefaultIfNil(a.AuthService, &auth.AuthService{})
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "debug":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			if err := parser.ParseBool(d, &a.ServerOptions.Debug); err != nil {
				return err
			}
		case "trace":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			if err := parser.ParseBool(d, &a.ServerOptions.Trace); err != nil {
				return err
			}
			if a.ServerOptions.Trace {
				a.ServerOptions.Debug = true
			}
		case "cluster":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Cluster = fnutils.DefaultIfNil(a.ServerOptions.Cluster, &embedded.Cluster{})
			if err := natscaddyfile.ParseCluster(d, a.ServerOptions.Cluster); err != nil {
				return err
			}
		case "auth", "authorization":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Authorization = fnutils.DefaultIfNil(a.ServerOptions.Authorization, &embedded.AuthorizationMap{})
			if err := natscaddyfile.ParseAuthorization(d, a.ServerOptions.Authorization); err != nil {
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
			a.AuthService.DefaultHandlerRaw = caddyconfig.JSONModuleObject(mod, "module", name, nil)
		case "auth_service":
			a.AuthService = fnutils.DefaultIfNil(a.AuthService, &auth.AuthService{})
			if err := a.AuthService.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		case "accounts":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.Connectors = fnutils.DefaultIfEmpty(a.Connectors, client.Connections{})
			a.ServerOptions.Accounts = fnutils.DefaultIfEmpty(a.ServerOptions.Accounts, []*embedded.Account{})
			if err := natscaddyfile.ParseAccounts(d, a.AuthService, &a.Connectors, &a.ServerOptions.Accounts); err != nil {
				return err
			}
		case "account":
			acc := embedded.Account{}
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.Connectors = fnutils.DefaultIfEmpty(a.Connectors, client.Connections{})
			a.ServerOptions.Accounts = fnutils.DefaultIfEmpty(a.ServerOptions.Accounts, []*embedded.Account{})
			if err := parser.ParseString(d, &acc.Name); err != nil {
				return err
			}
			if err := natscaddyfile.ParseAccount(d, a.AuthService, &a.Connectors, &acc); err != nil {
				return err
			}
			a.ServerOptions.Accounts = append(a.ServerOptions.Accounts, &acc)
		case "jetstream":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.JetStream = fnutils.DefaultIfNil(a.ServerOptions.JetStream, &embedded.JetStream{})
			if err := natscaddyfile.ParseJetStream(d, a.ServerOptions.JetStream); err != nil {
				return err
			}
		case "nats_server", "server":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			if err := natscaddyfile.ParseOptions(d, a.ServerOptions); err != nil {
				return err
			}
		case "mqtt_server":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Mqtt = fnutils.DefaultIfNil(a.ServerOptions.Mqtt, &embedded.MQTT{})
			if err := natscaddyfile.ParseMqtt(d, a.ServerOptions.Mqtt); err != nil {
				return err
			}
		case "websocket_server":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Websocket = fnutils.DefaultIfNil(a.ServerOptions.Websocket, &embedded.Websocket{})
			if err := natscaddyfile.ParseWebsocket(d, a.ServerOptions.Websocket); err != nil {
				return err
			}
		case "remote_server", "remote_connection", "leafnode_connection":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Leafnode = fnutils.DefaultIfNil(a.ServerOptions.Leafnode, &embedded.Leafnode{})
			a.ServerOptions.Leafnode.Remotes = fnutils.DefaultIfEmpty(a.ServerOptions.Leafnode.Remotes, []*embedded.Remote{})
			remote := embedded.Remote{}
			if err := natscaddyfile.ParseRemoteLeafnode(d, &remote); err != nil {
				return err
			}
			a.ServerOptions.Leafnode.Remotes = append(a.ServerOptions.Leafnode.Remotes, &remote)
		case "leafnode", "leafnode_server":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Leafnode = fnutils.DefaultIfNil(a.ServerOptions.Leafnode, &embedded.Leafnode{})
			if err := natscaddyfile.ParseLeafnodes(d, a.ServerOptions.Leafnode); err != nil {
				return err
			}
		case "metrics":
			a.ServerOptions = fnutils.DefaultIfNil(a.ServerOptions, embedded.NewOptions())
			a.ServerOptions.Metrics = fnutils.DefaultIfNil(a.ServerOptions.Metrics, &embedded.Metrics{})
			if err := natscaddyfile.ParseMetrics(d, a.ServerOptions.Metrics); err != nil {
				return err
			}
		case "ready_timeout":
			if err := parser.ParseDuration(d, &a.ReadyTimeout); err != nil {
				return err
			}
		case "connector":
			connector := client.Connection{}
			if err := natscaddyfile.ParseConnector(d, &connector); err != nil {
				return err
			}
			a.Connectors = fnutils.DefaultIfEmpty(a.Connectors, client.Connections{})
			a.Connectors = append(a.Connectors, &connector)
		default:
			return d.Errf("unknown directive '%s'", d.Val())
		}
	}
	// Remove empty auth service
	if a.AuthService.Zero() {
		a.AuthService = nil
	}
	// Remove empty connectors
	if len(a.Connectors) == 0 {
		a.Connectors = nil
	}
	return nil
}
