package natsmagic

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

func ParseNatsOptions(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	app := new(App)
	app.NATS = new(NatsConfig)
	app.Policies = new(AppPolicies)
	app.Policies.StandardPolicies = caddytls.ConnectionPolicies{}
	app.Policies.WebsocketPolicies = caddytls.ConnectionPolicies{}
	app.Policies.LeafnodePolicies = caddytls.ConnectionPolicies{}
	app.Policies.MQTTPolicies = caddytls.ConnectionPolicies{}
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing nats values of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, app)
		if err != nil {
			return nil, err
		}
	}

	err := app.UnmarshalCaddyfile(d)

	return httpcaddyfile.App{
		Name:  "nats",
		Value: caddyconfig.JSON(app, nil),
	}, err
}

func (app *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	opts := app.NATS
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "sni":
				if !d.NextArg() {
					return d.ArgErr()
				}
				domains := []string{}
				domains = append(domains, d.Val())
				for d.NextArg() {
					domains = append(domains, d.Val())
				}
				policy := &caddytls.ConnectionPolicy{
					MatchersRaw: caddy.ModuleMap{
						"sni": caddyconfig.JSON(domains, nil),
					},
				}
				app.Policies.StandardPolicies = append(app.Policies.StandardPolicies, policy)
				opts.SNI = domains[0]
			case "no_tls":
				opts.NoTLS = true
			case "config_file":
				if !d.AllArgs(&opts.ConfigFile) {
					return d.ArgErr()
				}
			case "server_name":
				if !d.AllArgs(&opts.ServerName) {
					return d.ArgErr()
				}
			case "server_tags":
				for d.NextArg() {
					opts.ServerTags = append(opts.ServerTags, d.Val())
				}
				if len(opts.ServerTags) == 0 {
					return d.Err("empty server tags")
				}
			case "host":
				if !d.AllArgs(&opts.Host) {
					return d.ArgErr()
				}
			case "port":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid port: %v", err)
				}
				opts.Port = t
			case "client_advertise":
				if !d.AllArgs(&opts.ClientAdvertise) {
					return d.ArgErr()
				}
			case "debug":
				opts.Debug = true
			case "trace":
				opts.Trace = true
			case "trace_verbose":
				opts.TraceVerbose = true
			case "http_port":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid http_port: %v", err)
				}
				opts.HTTPPort = t
			case "https_port":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid https_port: %v", err)
				}
				opts.HTTPSPort = t
			case "http_base_path":
				if !d.AllArgs(&opts.HTTPBasePath) {
					return d.ArgErr()
				}
			case "disable_logging":
				opts.NoLog = true
			case "disable_sublist_cache":
				opts.NoSublistCache = true
			case "max_connections":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid max_connections: %v", err)
				}
				opts.MaxConn = t
			case "max_payload":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid max_payload: %v", err)
				}
				opts.MaxPayload = t
			case "max_pending":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid max_pending: %v", err)
				}
				opts.MaxPending = t
			case "max_subscriptions":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid max_subscriptions: %v", err)
				}
				opts.MaxSubs = t
			case "max_control_line":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid max_control_line: %v", err)
				}
				opts.MaxControlLine = t
			case "ping_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid ping_interval: %v", err)
				}
				opts.PingInterval = time.Duration(t) * time.Second
			case "ping_max":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid ping_max: %v", err)
				}
				opts.PingMax = t
			case "write_deadline":
				if !d.NextArg() {
					return d.ArgErr()
				}
				t, err := strconv.Atoi(d.Val())
				if err != nil {
					return d.Errf("invalid write_deadline: %v", err)
				}
				opts.WriteDeadline = time.Duration(t) * time.Second
			case "no_auth_user":
				if !d.AllArgs(&opts.NoAuthUser) {
					return d.ArgErr()
				}
			case "jetstream":
				opts.JetStream = &JetStream{}
				if d.NextArg() {
					opts.JetStream.StoreDir = d.Val()
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						return d.ArgErr()
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "domain":
							if !d.AllArgs(&opts.JetStream.Domain) {
								return d.ArgErr()
							}
						case "store_dir":
							if !d.AllArgs(&opts.JetStream.StoreDir) {
								return d.ArgErr()
							}
						case "max_memory":
							if !d.NextArg() {
								return d.ArgErr()
							}
							t, err := strconv.Atoi(d.Val())
							if err != nil {
								return d.Errf("invalid max_memory: %v", err)
							}
							opts.JetStream.MaxMemory = t
						case "max_file":
							if !d.NextArg() {
								return d.ArgErr()
							}
							t, err := strconv.Atoi(d.Val())
							if err != nil {
								return d.Errf("invalid max_file: %v", err)
							}
							opts.JetStream.MaxFile = t
						default:
							return d.Errf("unrecognized subdirective: %s", d.Val())
						}
					}
				}
			case "mqtt":
				if d.NextArg() {
					port, err := strconv.Atoi(d.Val())
					if err != nil {
						return d.Errf("invalid mqtt port: %v", err)
					}
					opts.MQTT = &MQTT{Port: port}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						return d.ArgErr()
					}
				} else {
					opts.MQTT = &MQTT{}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "host":
							if !d.AllArgs(&opts.MQTT.Host) {
								return d.ArgErr()
							}
						case "port":
							if !d.NextArg() {
								return d.ArgErr()
							}
							t, err := strconv.Atoi(d.Val())
							if err != nil {
								return d.Errf("invalid mqtt port: %v", err)
							}
							opts.MQTT.Port = t
						case "jetstream_domain":
							if !d.AllArgs(&opts.MQTT.JSDomain) {
								return d.ArgErr()
							}
						case "no_tls":
							opts.MQTT.NoTLS = true
						case "sni":
							if !d.NextArg() {
								return d.ArgErr()
							}
							domains := []string{}
							domains = append(domains, d.Val())
							for d.NextArg() {
								domains = append(domains, d.Val())
							}
							policy := &caddytls.ConnectionPolicy{
								MatchersRaw: caddy.ModuleMap{
									"sni": caddyconfig.JSON(domains, nil),
								},
							}
							app.Policies.MQTTPolicies = append(app.Policies.MQTTPolicies, policy)
						default:
							return d.Errf("unrecognized subdirective: %s", d.Val())
						}
					}
				}
			case "websocket":
				if d.NextArg() {
					port, err := strconv.Atoi(d.Val())
					if err != nil {
						return d.Errf("invalid websocket port: %v", err)
					}
					opts.Websocket = &Websocket{Port: port}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						return d.ArgErr()
					}
				} else {
					opts.Websocket = &Websocket{}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "host":
							if !d.AllArgs(&opts.Websocket.Host) {
								return d.ArgErr()
							}
						case "port":
							if !d.NextArg() {
								return d.ArgErr()
							}
							t, err := strconv.Atoi(d.Val())
							if err != nil {
								return d.Errf("invalid websocket port: %v", err)
							}
							opts.Websocket.Port = t
						case "advertise":
							if !d.AllArgs(&opts.Websocket.Advertise) {
								return d.ArgErr()
							}
						case "sni":
							if !d.NextArg() {
								return d.ArgErr()
							}
							domains := []string{}
							domains = append(domains, d.Val())
							for d.NextArg() {
								domains = append(domains, d.Val())
							}
							policy := &caddytls.ConnectionPolicy{
								MatchersRaw: caddy.ModuleMap{
									"sni": caddyconfig.JSON(domains, nil),
								},
							}
							app.Policies.WebsocketPolicies = append(app.Policies.WebsocketPolicies, policy)
						case "no_tls":
							opts.Websocket.NoTLS = true
						default:
							return d.Errf("unrecognized subdirective: %s", d.Val())
						}
					}
				}
			case "leafnodes":
				opts.LeafNode = &LeafNode{}
				if d.NextArg() {
					port, err := strconv.Atoi(d.Val())
					if err != nil {
						return d.Errf("invalid leafnodes port: %v", err)
					}
					opts.LeafNode.Port = port
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						return d.ArgErr()
					}
				} else {
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "host":
							if !d.AllArgs(&opts.LeafNode.Host) {
								return d.ArgErr()
							}
						case "port":
							if !d.NextArg() {
								return d.ArgErr()
							}
							t, err := strconv.Atoi(d.Val())
							if err != nil {
								return d.Errf("invalid leafnode port: %v", err)
							}
							opts.LeafNode.Port = t
						case "advertise":
							if !d.AllArgs(&opts.LeafNode.Advertise) {
								return d.ArgErr()
							}
						case "sni":
							if !d.NextArg() {
								return d.ArgErr()
							}
							domains := []string{}
							domains = append(domains, d.Val())
							for d.NextArg() {
								domains = append(domains, d.Val())
							}
							policy := &caddytls.ConnectionPolicy{
								MatchersRaw: caddy.ModuleMap{
									"sni": caddyconfig.JSON(domains, nil),
								},
							}
							app.Policies.LeafnodePolicies = append(app.Policies.LeafnodePolicies, policy)
						case "no_tls":
							opts.LeafNode.NoTLS = true
						case "remotes":
							opts.LeafNode.Remotes = []Remote{}
							for nesting := d.Nesting(); d.NextBlock(nesting); {
								remote := Remote{Url: d.Val()}
								if !d.NextArg() {
									opts.LeafNode.Remotes = append(opts.LeafNode.Remotes, remote)
								} else {
									for nesting := d.Nesting(); d.NextBlock(nesting); {
										switch d.Val() {
										case "url":
											if !d.AllArgs(&remote.Url) {
												return d.ArgErr()
											}
										case "account":
											if !d.AllArgs(&remote.Account) {
												return d.ArgErr()
											}
										case "credentials":
											if !d.AllArgs(&remote.Credentials) {
												return d.ArgErr()
											}
										default:
											return d.Errf("unrecognized subdirective: %s", d.Val())
										}
									}
									opts.LeafNode.Remotes = append(opts.LeafNode.Remotes, remote)
								}
							}
						}
					}
				}
			case "operator":
				if !d.AllArgs(&opts.Operator) {
					return d.ArgErr()
				}
			case "system_account":
				if !d.AllArgs(&opts.SystemAccount) {
					return d.ArgErr()
				}
			case "resolver_preload":
				opts.ResolverPreload = []string{}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					if !d.NextArg() {
						return d.ArgErr()
					}
					opts.ResolverPreload = append(opts.ResolverPreload, d.Val())
				}
			case "resolver":
				opts.Resolver = &AccountResolver{}
				if d.NextArg() {
					switch d.Val() {
					case "memory":
						opts.Resolver.Memory = true
						for nesting := d.Nesting(); d.NextBlock(nesting); {
							return d.ArgErr()
						}
					case "full":
						opts.Resolver.Full = true
						for nesting := d.Nesting(); d.NextBlock(nesting); {
							switch d.Val() {
							case "path":
								if !d.AllArgs(&opts.Resolver.Path) {
									return d.ArgErr()
								}
							default:
								return d.Errf("unrecognized subdirective: %s", d.Val())
							}
						}
					case "cache":
						opts.Resolver.Full = false
						for nesting := d.Nesting(); d.NextBlock(nesting); {
							switch d.Val() {
							case "path":
								if !d.AllArgs(&opts.Resolver.Path) {
									return d.ArgErr()
								}
							default:
								return d.Errf("unrecognized subdirective: %s", d.Val())
							}
						}
					default:
						return d.Errf("unrecognized subdirective: %s", d.Val())
					}
				}
			case "metrics":
				opts.Metrics = &Metrics{}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "host":
						if !d.AllArgs(&opts.Metrics.Host) {
							return d.ArgErr()
						}
					case "port":
						if !d.NextArg() {
							return d.ArgErr()
						}
						t, err := strconv.Atoi(d.Val())
						if err != nil {
							return d.Errf("invalid metrics port: %v", err)
						}
						opts.Metrics.Port = t
					case "base_path":
						if !d.AllArgs(&opts.Metrics.BasePath) {
							return d.ArgErr()
						}
					}
				}
			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}
	return nil
}
