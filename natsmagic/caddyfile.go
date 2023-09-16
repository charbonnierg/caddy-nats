package natsmagic

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

func ParseNatsOptions(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	app := new(App)
	app.Options = new(Options)
	app.StandardPolicies = caddytls.ConnectionPolicies{}

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
	opts := app.Options
	allDomains := [][]string{}
	domains := []string{}
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "sni":
				if !d.NextArg() {
					return d.ArgErr()
				}
				domains = append(domains, d.Val())
				for d.NextArg() {
					domains = append(domains, d.Val())
				}
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

			default:
				return d.Errf("unrecognized subdirective: %s", d.Val())
			}
		}
	}
	automations := make([]caddytls.AutomationPolicy, len(allDomains))
	allDomains = append(allDomains, domains)
	for _, domains := range allDomains {
		automation := caddytls.AutomationPolicy{
			SubjectsRaw: domains,
		}
		automations = append(automations, automation)
	}
	app.Automations = automations
	return nil
}
