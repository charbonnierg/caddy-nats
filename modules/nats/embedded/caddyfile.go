// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package embedded

import (
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func (o *Options) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {

	// Do not expect any argument but o block instead
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "tls":
			if err := parseTLS(d, o.TLS); err != nil {
				return err
			}
		case "no_tls":
			if err := parseNoTLS(d, o); err != nil {
				return err
			}
		case "server_name":
			if !d.AllArgs(&o.ServerName) {
				return d.Err("server_name requires exactly one name value")
			}
		// Alias for server_name
		case "name":
			if !d.AllArgs(&o.ServerName) {
				return d.Err("name requires exactly one name value")
			}
		case "server_tags":
			if err := parseServerTags(d, o); err != nil {
				return err
			}
		// Alias for "server_tags"
		case "tags":
			if err := parseServerTags(d, o); err != nil {
				return err
			}
		case "host":
			if !d.AllArgs(&o.Host) {
				return d.Err("host requires exactly one listen address")
			}
		case "port":
			if err := parsePort(d, o); err != nil {
				return err
			}
		case "advertise":
			if !d.AllArgs(&o.Advertise) {
				return d.Err("advertise requires exactly one address (including port, but not scheme)")
			}
		case "debug":
			if err := parseDebug(d, o); err != nil {
				return err
			}
		case "trace":
			if err := parseTrace(d, o); err != nil {
				return err
			}
		case "trace_verbose":
			if err := parseTraceVerbose(d, o); err != nil {
				return err
			}
		case "http_port":
			if err := parseHttpPort(d, o); err != nil {
				return err
			}
		case "http_host":
			if !d.AllArgs(&o.HTTPHost) {
				return d.Err("invalid http_host option")
			}
		case "https_port":
			if err := parseHttpsPort(d, o); err != nil {
				return err
			}
		case "http_base_path":
			if !d.AllArgs(&o.HTTPBasePath) {
				return d.Err("http_base_path requires exactly one path")
			}
		case "disable_logging":
			if err := parseDisableLogging(d, o); err != nil {
				return err
			}
		case "disable_sublist_cache":
			if err := parseDisableSublistCache(d, o); err != nil {
				return err
			}
		case "max_connections":
			if err := parseMaxConnections(d, o); err != nil {
				return err
			}
		case "max_payload":
			if err := parseMaxPayload(d, o); err != nil {
				return err
			}
		case "max_pending":
			if err := parseMaxPending(d, o); err != nil {
				return err
			}
		case "max_subscriptions":
			if err := parseMaxSubscriptions(d, o); err != nil {
				return err
			}
		case "max_control_line":
			if err := parseMaxControlLine(d, o); err != nil {
				return err
			}
		case "ping_interval":
			if err := parsePingInterval(d, o); err != nil {
				return err
			}
		case "ping_max":
			if err := parsePingMax(d, o); err != nil {
				return err
			}
		case "write_deadline":
			if err := parseWriteDeadline(d, o); err != nil {
				return err
			}
		case "no_auth_user":
			if !d.AllArgs(&o.NoAuthUser) {
				return d.ArgErr()
			}
		case "jetstream":
			if err := parseJetStream(d, o); err != nil {
				return err
			}
		case "mqtt":
			if err := parseMqtt(d, o); err != nil {
				return err
			}
		case "websocket":
			if err := parseWebsocket(d, o); err != nil {
				return err
			}
		case "leafnodes":
			if err := parseLeafnodes(d, o); err != nil {
				return err
			}
		case "operator":
			if err := parseOperator(d, o); err != nil {
				return err
			}
		case "system_account":
			if !d.AllArgs(&o.SystemAccount) {
				return d.Err("system_account requires exactly one account name")
			}
		case "accounts":
			if err := parseAccounts(d, o); err != nil {
				return err
			}
		case "users":
			if err := parseUsers(d, o); err != nil {
				return err
			}
			// case "resolver":
		// 	if err := parseResolver(d, o); err != nil {
		// 		return err
		// 	}
		case "metrics":
			if err := parseMetrics(d, o); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized nats_server subdirective: %s", d.Val())
		}
	}
	// Listen on localhost only by default if not configured otherwise
	if o.HTTPPort != 0 && o.HTTPHost == "" {
		o.HTTPHost = "127.0.0.1"
	}
	return nil
}

func parseAccounts(d *caddyfile.Dispenser, o *Options) error {
	if o.Accounts == nil {
		o.Accounts = []*Account{}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		acc := Account{Name: name}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "jetstream":
				acc.JetStream = true
			}
		}
		o.Accounts = append(o.Accounts, &acc)
	}
	return nil
}

func parseUsers(d *caddyfile.Dispenser, o *Options) error {
	if o.Authorization == nil {
		o.Authorization = &AuthorizationMap{
			Users: []User{},
		}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		user := User{User: name}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "password":
				if !d.AllArgs(&user.Password) {
					return d.Err("password requires exactly one password value")
				}
			default:
				return d.Errf("unrecognized user subdirective: %s", d.Val())
			}
		}
		o.Authorization.Users = append(o.Authorization.Users, user)
	}
	return nil
}

// parseServerTags parses the "tags" directive found in o Caddyfile "nats_server" option block.
func parseServerTags(d *caddyfile.Dispenser, o *Options) error {
	tags := d.RemainingArgs()
	if len(tags) == 0 {
		return d.Err("tags requires at least one tag value")
	}
	validTags := map[string]string{}
	for _, tag := range tags {
		if len(tag) == 0 {
			return d.Err("empty tag value")
		}
		keyvalue := strings.Split(tag, ":")
		if len(keyvalue) != 2 {
			return d.Err("invalid tag value")
		}
		key := strings.TrimSpace(keyvalue[0])
		value := strings.TrimSpace(keyvalue[1])
		if len(key) == 0 || len(value) == 0 {
			return d.Err("empty tag key or value")
		}
		if _, ok := validTags[key]; ok {
			return d.Err("duplicate tag key")
		}
		validTags[key] = value
	}
	o.ServerTags = validTags
	return nil
}

// parseNoTLS parses the "no_tls" directive found in o Caddyfile "nats_server" option block.
func parseNoTLS(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("no_tls does not take any argument")
	}
	o.NoTLS = true
	return nil
}

// parseDisableLogging parses the "disable_logging" directive found in o Caddyfile "nats_server" option block.
func parseDisableLogging(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("disable_logging does not take any argument")
	}
	o.NoLog = true
	return nil
}

// parseDisableSublistCache parses the "disable_sublist_cache" directive found in o Caddyfile "nats_server" option block.
func parseDisableSublistCache(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("disable_sublist_cache does not take any argument")
	}
	o.NoSublistCache = true
	return nil
}

// parseDebug parses the "debug" directive found in o Caddyfile "nats_server" option block.
func parseDebug(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("debug does not take any argument")
	}
	o.Debug = true
	return nil
}

// parseTrace parses the "trace" directive found in o Caddyfile "nats_server" option block.
func parseTrace(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("trace does not take any argument")
	}
	o.Debug = true
	o.Trace = true
	return nil
}

// parseTraceVerbose parses the "trace_verbose" directive found in o Caddyfile "nats_server" option block.
func parseTraceVerbose(d *caddyfile.Dispenser, o *Options) error {
	if d.NextArg() {
		return d.Err("trace_verbose does not take any argument")
	}
	o.Debug = true
	o.Trace = true
	o.TraceVerbose = true
	return nil
}

// parsePort parses the "port" directive found in o Caddyfile "nats_server" option block.
func parsePort(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("port requires exactly one port number")
	}
	port, err := parseutils.ParsePort(raw)
	if err != nil {
		return d.Err(err.Error())
	}
	o.Port = port
	return nil
}

// parseHttpPort parses the "http_port" directive found in o Caddyfile "nats_server" option block.
func parseHttpPort(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("http_port requires exactly one port number")
	}
	port, err := parseutils.ParsePort(raw)
	if err != nil {
		return d.Err(err.Error())
	}
	o.HTTPPort = port
	return nil
}

// parseHttpsPort parses the "https_port" directive found in o Caddyfile "nats_server" option block.
func parseHttpsPort(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("https_port requires exactly one port number")
	}
	port, err := parseutils.ParsePort(raw)
	if err != nil {
		return d.Err(err.Error())
	}
	o.HTTPSPort = port
	return nil
}

// parseMaxConnections parses the "max_connections" directive found in o Caddyfile "nats_server" option block.
func parseMaxConnections(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("max_connections requires exactly one integer value")
	}
	t, err := strconv.Atoi(raw)
	if err != nil {
		return d.Errf("invalid max_connections: %v", err)
	}
	o.MaxConn = t
	return nil
}

// parseMaxPayload parses the "max_payload" directive found in o Caddyfile "nats_server" option block.
func parseMaxPayload(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("max_payload requires exactly one size value")
	}
	size, err := parseutils.ParseBytes(raw)
	if err != nil {
		return d.Errf("invalid max_payload: %s", err.Error())
	}
	size32, err := parseutils.Int32(size)
	if err != nil {
		return d.Errf("invalid max_payload: %s", err.Error())
	}
	o.MaxPayload = size32
	return nil
}

// parseMaxPending parses the "max_pending" directive found in o Caddyfile "nats_server" option block.
func parseMaxPending(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("max_pending requires exactly one size value")
	}
	size, err := parseutils.ParseBytes(raw)
	if err != nil {
		return d.Errf("invalid max_pending: %s", err.Error())
	}
	o.MaxPending = int64(size)
	return nil
}

// parseMaxControlLine parses the "max_control_line" directive found in o Caddyfile "nats_server" option block.
func parseMaxControlLine(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("max_control_line requires exactly one size value")
	}
	size, err := parseutils.ParseBytes(raw)
	if err != nil {
		return d.Errf("invalid max_control_line: %s", err.Error())
	}
	size32, err := parseutils.Int32(size)
	if err != nil {
		return d.Errf("invalid max_control_line: %s", err.Error())
	}
	o.MaxControlLine = size32
	return nil
}

// parseMaxSubscriptions parses the "max_subscriptions" directive found in o Caddyfile "nats_server" option block.
func parseMaxSubscriptions(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("max_subscriptions requires exactly one integer value")
	}
	count, err := strconv.Atoi(raw)
	if err != nil {
		return d.Errf("invalid max_subscriptions: %v", err)
	}
	o.MaxSubs = count
	return nil
}

// parsePingInterval parses the "ping_interval" directive found in o Caddyfile "nats_server" option block.
func parsePingInterval(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("ping_interval requires exactly one duration value")
	}
	duration, err := caddy.ParseDuration(raw)
	if err != nil {
		return d.Errf("invalid ping_interval: %v", err)
	}
	o.PingInterval = duration
	return nil
}

// parsePingMax parses the "ping_max" directive found in o Caddyfile "nats_server" option block.
func parsePingMax(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("ping_max requires exactly one integer value")
	}
	max, err := strconv.Atoi(d.Val())
	if err != nil {
		return d.Errf("invalid ping_max: %v", err)
	}
	o.MaxPingsOut = max
	return nil
}

// parseWriteDeadline parses the "write_deadline" directive found in o Caddyfile "nats_server" option block.
func parseWriteDeadline(d *caddyfile.Dispenser, o *Options) error {
	raw := ""
	if !d.AllArgs(&raw) {
		return d.Err("write_deadline requires exactly one duration value")
	}
	duration, err := caddy.ParseDuration(raw)
	if err != nil {
		return d.Errf("invalid write_deadline: %v", err)
	}
	o.WriteDeadline = duration
	return nil
}

// parseJetStream parses the "jetstream" directive found in o Caddyfile "nats_server" option block.
func parseJetStream(d *caddyfile.Dispenser, o *Options) error {
	// Make sure we have o JetStream config
	if o.JetStream == nil {
		o.JetStream = &JetStream{}
	}
	jsopts := o.JetStream
	// short-syntax
	if d.NextArg() {
		jsopts.StoreDir = d.Val()
		if d.NextArg() {
			return d.Err("jetstream short syntax requires exactly one path to store directory")
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("jetstream short syntax requires exactly one path to store directory")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "domain":
				if !d.AllArgs(&jsopts.Domain) {
					return d.Err("jetstream.domain requires exactly one domain name")
				}
			case "store_dir":
				if !d.AllArgs(&jsopts.StoreDir) {
					return d.Err("jetstream.store_dir requires exactly one path")
				}
			case "max_memory":
				raw := ""
				if !d.AllArgs(&raw) {
					return d.Err("jetstream.max_memory requires exactly one size value")
				}
				size, err := parseutils.ParseBytes(raw)
				if err != nil {
					return d.Errf("invalid jetstream.max_memory: %s", err.Error())
				}
				jsopts.MaxMemory = int64(size)
			case "max_file":
				raw := ""
				if !d.AllArgs(&raw) {
					return d.Err("jetstream.max_file requires exactly one size value")
				}
				size, err := parseutils.ParseBytes(raw)
				if err != nil {
					return d.Errf("invalid jetstream.max_file: %s", err.Error())
				}
				jsopts.MaxFile = int64(size)
			default:
				return d.Errf("unrecognized jetstream subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseMqtt parses the "mqtt" directive found in o Caddyfile "nats_server" option block.
func parseMqtt(d *caddyfile.Dispenser, o *Options) error {
	// Make sure we have o MQTT config
	if o.MQTT == nil {
		o.MQTT = &MQTT{}
	}
	mqttopts := o.MQTT
	// Short syntax
	if d.NextArg() {
		port, err := parseutils.ParsePort(d.Val())
		if err != nil {
			return d.Errf("invalid mqtt port: %v", err)
		}
		mqttopts.Port = port
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("mqtt short syntax requires exactly one port number")
		}
		// Long syntax
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "host":
				if !d.AllArgs(&mqttopts.Host) {
					return d.ArgErr()
				}
			case "port":
				raw := ""
				if !d.AllArgs(&raw) {
					return d.Err("mqtt.port requires exactly one port number")
				}
				port, err := parseutils.ParsePort(raw)
				if err != nil {
					return d.Errf("invalid mqtt port: %v", err)
				}
				mqttopts.Port = port
			case "jetstream_domain":
				if !d.AllArgs(&mqttopts.JSDomain) {
					return d.Err("mqtt.jetstream_domain requires exactly one value")
				}
			case "no_tls":
				if d.NextArg() {
					return d.Err("mqtt.no_tls does not take any argument")
				}
				mqttopts.NoTLS = true
			case "tls":
				if err := parseTLS(d, mqttopts.TLS); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized mqtt subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseWebsocket parses the "websocket" directive found in o Caddyfile "nats_server" option block.
func parseWebsocket(d *caddyfile.Dispenser, o *Options) error {
	// Make sure we have o Websocket config
	if o.Websocket == nil {
		o.Websocket = &Websocket{}
	}
	if d.NextArg() {
		port, err := parseutils.ParsePort(d.Val())
		if err != nil {
			return d.Errf("invalid websocket port: %v", err)
		}
		o.Websocket.Port = port
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("websocket short syntax requires exactly one port number")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "host":
				if !d.AllArgs(&o.Websocket.Host) {
					return d.ArgErr()
				}
			case "port":
				raw := ""
				if !d.AllArgs(&raw) {
					return d.Err("websocket.port requires exactly one port number")
				}
				port, err := parseutils.ParsePort(raw)
				if err != nil {
					return d.Errf("invalid websocket port: %v", err)
				}
				o.Websocket.Port = port
			case "advertise":
				if !d.AllArgs(&o.Websocket.Advertise) {
					return d.Err("websocket.advertise requires exactly one address (including port, but not scheme)")
				}
			case "tls":
				if o.Websocket.TLS == nil {
					o.Websocket.TLS = &TLSMap{}
				}
				if err := parseTLS(d, o.Websocket.TLS); err != nil {
					return err
				}
			case "no_tls":
				o.Websocket.NoTLS = true
			default:
				return d.Errf("unrecognized websocket subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseLeafnodes parse the "leafnodes" directive found in o Caddyfile "nats_server" option block.
func parseLeafnodes(d *caddyfile.Dispenser, o *Options) error {
	// Make sure we have o LeafNode config
	if o.Leafnode == nil {
		o.Leafnode = &Leafnode{}
	}
	leafopts := o.Leafnode
	// Short syntax
	if d.NextArg() {
		port, err := parseutils.ParsePort(d.Val())
		if err != nil {
			return d.Errf("invalid leafnodes port: %v", err)
		}
		leafopts.Port = port
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("leafnodes short syntax requires exactly one port number")
		}
		// Long syntax
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "host":
				if !d.AllArgs(&leafopts.Host) {
					return d.ArgErr()
				}
			case "port":
				raw := ""
				if !d.AllArgs(&raw) {
					return d.Err("leafnodes.port requires exactly one port number")
				}
				port, err := parseutils.ParsePort(raw)
				if err != nil {
					return d.Errf("invalid leafnode port: %v", err)
				}
				leafopts.Port = port
			case "advertise":
				if !d.AllArgs(&leafopts.Advertise) {
					return d.Err("leafnodes.advertise requires exactly one address (including port, but not scheme)")
				}
			case "tls":
				if err := parseTLS(d, leafopts.TLS); err != nil {
					return err
				}
			case "no_tls":
				if d.NextArg() {
					return d.Err("leafnodes.no_tls does not take any argument")
				}
				leafopts.NoTLS = true
			case "remotes":
				if err := parseRemoteLeafnodes(d, leafopts); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// parseRemoteLeafnodes parse the "remote_leafnodes" directive found in o Caddyfile "nats_server" option block.
func parseRemoteLeafnodes(d *caddyfile.Dispenser, leafopts *Leafnode) error {
	leafopts.Remotes = []Remote{}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		remote := Remote{Url: d.Val()}
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
		leafopts.Remotes = append(leafopts.Remotes, remote)
	}
	return nil
}

// parseOperator parses the "operator" directive found in o Caddyfile "nats_server" option block.
func parseOperator(d *caddyfile.Dispenser, o *Options) error {
	if o.Operators == nil {
		o.Operators = []string{}
	}
	op := ""
	if !d.AllArgs(&op) {
		return d.Err("operator requires exactly one operator name")
	}
	for _, existing := range o.Operators {
		if existing == op {
			return nil
		}
	}
	o.Operators = append(o.Operators, op)
	return nil
}

// parseResolver parses the "resolver" directive found in o Caddyfile "nats_server" option block.
// func parseResolver(d *caddyfile.Dispenser, o *Options) error {
// 	if !d.NextArg() {
// 		return d.Err("resolver requires exactly one resolver type followed by optional subdirectives")
// 	}
// 	switch d.Val() {
// 	case "memory":
// 		resopts.Memory = true
// 		if d.NextArg() {
// 			return d.Err("memory resolver does not take any argument")
// 		}
// 		for nesting := d.Nesting(); d.NextBlock(nesting); {
// 			return d.Err("memory resolver does not take any subdirective")
// 		}
// 	case "full":
// 		// Short syntax
// 		if d.NextArg() {
// 			resopts.Path = d.Val()
// 			if d.NextArg() {
// 				return d.Err("full resolver short syntax requires exactly one path")
// 			}
// 			for nesting := d.Nesting(); d.NextBlock(nesting); {
// 				return d.Err("full resolver short syntax requires exactly one path")
// 			}
// 		} else {
// 			for nesting := d.Nesting(); d.NextBlock(nesting); {
// 				switch d.Val() {
// 				case "path":
// 					if !d.AllArgs(&resopts.Path) {
// 						return d.ArgErr()
// 					}
// 				default:
// 					return d.Errf("unrecognized full resolver subdirective: %s", d.Val())
// 				}
// 			}
// 		}
// 	case "cache":
// 		resopts.Cache = true
// 		// Short syntax
// 		if d.NextArg() {
// 			resopts.Path = d.Val()
// 			if d.NextArg() {
// 				return d.Err("cache resolver short syntax requires exactly one path")
// 			}
// 			for nesting := d.Nesting(); d.NextBlock(nesting); {
// 				return d.Err("cache resolver short syntax requires exactly one path")
// 			}
// 		} else {
// 			for nesting := d.Nesting(); d.NextBlock(nesting); {
// 				switch d.Val() {
// 				case "path":
// 					if !d.AllArgs(&resopts.Path) {
// 						return d.ArgErr()
// 					}
// 				default:
// 					return d.Errf("unrecognized cache resolver subdirective: %s", d.Val())
// 				}
// 			}
// 		}
// 	default:
// 		return d.Errf("unrecognized resolver subdirective: %s", d.Val())
// 	}
// 	return nil
// }

// parseTLS parses the "tls" directive found in o Caddyfile "nats_server" option block.
func parseTLS(d *caddyfile.Dispenser, tlsOpts *TLSMap) error {
	domains := []string{}
	subjects := d.RemainingArgs()
	for _, subject := range subjects {
		if subject != "" {
			domains = append(domains, subject)
		}
	}
	if len(domains) > 0 {
		if tlsOpts.Subjects == nil {
			tlsOpts.Subjects = []string{}
		}
		tlsOpts.Subjects = append(tlsOpts.Subjects, domains...)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "subjects":
			domains := []string{}
			for _, arg := range d.RemainingArgs() {
				if arg != "" {
					domains = append(domains, arg)
				}
			}
			if len(domains) == 0 {
				return d.Err("tls.sni requires at least one domain name")
			}
			if tlsOpts.Subjects == nil {
				tlsOpts.Subjects = []string{}
			}
			tlsOpts.Subjects = append(tlsOpts.Subjects, domains...)
		case "cert_file":
			if !d.AllArgs(&tlsOpts.CertFile) {
				return d.ArgErr()
			}
		case "key_file":
			if !d.AllArgs(&tlsOpts.KeyFile) {
				return d.ArgErr()
			}
		case "ca_file":
			if !d.AllArgs(&tlsOpts.CaFile) {
				return d.ArgErr()
			}
		}
	}
	return nil
}

// parseMetrics parses the "metrics" directive found in o Caddyfile "nats_server" option block.
func parseMetrics(d *caddyfile.Dispenser, o *Options) error {
	// Make sure we have o Metrics config
	if o.Metrics == nil {
		o.Metrics = &Metrics{}
	}
	metricopts := o.Metrics
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "server_label":
			if !d.AllArgs(&metricopts.ServerLabel) {
				return d.ArgErr()
			}
		case "server_url":
			if !d.AllArgs(&metricopts.ServerUrl) {
				return d.ArgErr()
			}
		case "healthz":
			metricopts.Healthz = true
		case "connz":
			metricopts.Connz = true
			for d.NextArg() {
				switch d.Val() {
				case "detailed":
					metricopts.ConnzDetailed = true
				default:
					return d.Err("invalid metrics connz option")
				}
			}
		case "subz":
			metricopts.Subz = true
		case "routez":
			metricopts.Routez = true
		case "gatewayz":
			metricopts.Gatewayz = true
		case "leafz":
			metricopts.Leafz = true
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
