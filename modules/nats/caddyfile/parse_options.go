// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"github.com/quara-dev/beyond/pkg/natsutils/embedded"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

// ParseOptions parses the "nats_server" option block found in the Caddyfile.
func ParseOptions(d *caddyfile.Dispenser, o *embedded.Options) error {

	// Do not expect any argument but o block instead
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "no_tls":
			if err := caddyutils.ParseBool(d, &o.NoTLS); err != nil {
				return err
			}
		case "name", "server_name":
			if err := caddyutils.ParseString(d, &o.ServerName); err != nil {
				return err
			}
		case "tags", "server_tags":
			if err := caddyutils.ParseKeyValuePairs(d, &o.ServerTags, ":"); err != nil {
				return err
			}
		case "host", "server_host":
			if err := caddyutils.ParseString(d, &o.Host); err != nil {
				return err
			}
		case "port", "server_port":
			if err := caddyutils.ParsePort(d, &o.Port); err != nil {
				return err
			}
		case "advertise", "client_advertise":
			if err := caddyutils.ParseString(d, &o.Advertise); err != nil {
				return err
			}
		case "debug", "enable_debug":
			if err := caddyutils.ParseBool(d, &o.Debug); err != nil {
				return err
			}
		case "trace", "enable_trace":
			if err := caddyutils.ParseBool(d, &o.Trace); err != nil {
				return err
			}
			if o.Trace {
				o.Debug = true
			}
		case "trace_verbose", "enable_trace_verbose":
			if err := caddyutils.ParseBool(d, &o.TraceVerbose); err != nil {
				return err
			}
			if o.TraceVerbose {
				o.Debug = true
				o.Trace = true
			}
		case "http_port", "monitoring_port":
			if err := caddyutils.ParsePort(d, &o.HTTPPort); err != nil {
				return err
			}
		case "http_host", "monitoring_host":
			if err := caddyutils.ParseString(d, &o.HTTPHost); err != nil {
				return err
			}
		case "https_port", "monitoring_tls_port":
			if err := caddyutils.ParsePort(d, &o.HTTPSPort); err != nil {
				return err
			}
		case "http_base_path", "monitoring_base_path":
			if err := caddyutils.ParseString(d, &o.HTTPBasePath); err != nil {
				return err
			}
		case "no_log", "disable_logging":
			if err := caddyutils.ParseBool(d, &o.NoLog); err != nil {
				return err
			}
		case "no_sublist_cache", "disable_sublist_cache":
			if err := caddyutils.ParseBool(d, &o.NoSublistCache); err != nil {
				return err
			}
		case "max_conn", "max_connections":
			if err := caddyutils.ParseInt(d, &o.MaxConn); err != nil {
				return err
			}
		case "max_payload":
			if err := caddyutils.ParseByteSizeI32(d, &o.MaxPayload); err != nil {
				return err
			}
		case "max_pending":
			if err := caddyutils.ParseByteSizeI64(d, &o.MaxPending); err != nil {
				return err
			}
		case "max_subs", "max_subscriptions":
			if err := caddyutils.ParseInt(d, &o.MaxSubs); err != nil {
				return err
			}
		case "max_control_line":
			if err := caddyutils.ParseByteSizeI32(d, &o.MaxControlLine); err != nil {
				return err
			}
		case "ping_interval":
			if err := caddyutils.ParseDuration(d, &o.PingInterval); err != nil {
				return err
			}
		case "max_pings_out", "ping_max":
			if err := caddyutils.ParseInt(d, &o.MaxPingsOut); err != nil {
				return err
			}
		case "write_deadline":
			if err := caddyutils.ParseDuration(d, &o.WriteDeadline); err != nil {
				return err
			}
		case "no_auth_user":
			if err := caddyutils.ParseString(d, &o.NoAuthUser); err != nil {
				return err
			}
		case "system_account":
			if err := caddyutils.ParseString(d, &o.SystemAccount); err != nil {
				return err
			}
		case "tls":
			o.TLS = fnutils.DefaultIfNil(o.TLS, &embedded.TLSMap{})
			if err := parseTLS(d, o.TLS); err != nil {
				return err
			}
		case "jetstream":
			o.JetStream = fnutils.DefaultIfNil(o.JetStream, &embedded.JetStream{})
			if err := parseJetStream(d, o.JetStream); err != nil {
				return err
			}
		case "mqtt":
			o.Mqtt = fnutils.DefaultIfNil(o.Mqtt, &embedded.MQTT{})
			if err := parseMqtt(d, o.Mqtt); err != nil {
				return err
			}
		case "websocket":
			o.Websocket = fnutils.DefaultIfNil(o.Websocket, &embedded.Websocket{})
			if err := parseWebsocket(d, o.Websocket); err != nil {
				return err
			}
		case "leafnodes", "leafnode":
			o.Leafnode = fnutils.DefaultIfNil(o.Leafnode, &embedded.Leafnode{})
			if err := parseLeafnodes(d, o.Leafnode); err != nil {
				return err
			}
		case "operators", "operator":
			o.Operators = fnutils.DefaultIfEmpty(o.Operators, []string{})
			if err := caddyutils.ParseStringArray(d, &o.Operators, false); err != nil {
				return err
			}
		case "accounts":
			o.Accounts = fnutils.DefaultIfEmpty(o.Accounts, []*embedded.Account{})
			if err := parseAccounts(d, &o.Accounts); err != nil {
				return err
			}
		case "users":
			o.Authorization = fnutils.DefaultIfNil(o.Authorization, &embedded.AuthorizationMap{})
			o.Authorization.Users = fnutils.DefaultIfEmpty(o.Authorization.Users, []embedded.User{})
			if err := parseAuthUsers(d, o.Authorization); err != nil {
				return err
			}
		case "metrics":
			o.Metrics = fnutils.DefaultIfNil(o.Metrics, &embedded.Metrics{})
			if err := parseMetrics(d, o.Metrics); err != nil {
				return err
			}
		case "full_resolver":
			o.FullResolver = fnutils.DefaultIfNil(o.FullResolver, &embedded.FullAccountResolver{})
			if err := parseFullResolver(d, o.FullResolver); err != nil {
				return err
			}
		case "cache_resolver":
			o.CacheResolver = fnutils.DefaultIfNil(o.CacheResolver, &embedded.CacheAccountResolver{})
			if err := parseCacheResolver(d, o.CacheResolver); err != nil {
				return err
			}
		case "memory_resolver":
			o.MemoryResolver = fnutils.DefaultIfNil(o.MemoryResolver, &embedded.MemoryAccountResolver{})
			if err := parseMemoryResolver(d, o.MemoryResolver); err != nil {
				return err
			}
		case "resolver":
			return d.Err("resolver directive has been removed, use full_resolver, cache_resolver or memory_resolver instead")
		default:
			return d.Errf("unrecognized nats_server subdirective: %s", d.Val())
		}
	}
	return nil
}

func parseSubjectMapping(d *caddyfile.Dispenser, account *embedded.Account) error {
	if account == nil {
		return d.Err("internal error: account is nil. Please open a bug report.")
	}
	if account.Mappings == nil {
		return d.Err("internal error: mappings is nil. Please open a bug report.")
	}
	mapping := embedded.SubjectMapping{MapDest: []*server.MapDest{}}
	if err := caddyutils.ParseString(d, &mapping.Subject); err != nil {
		return err
	}
	if d.CountRemainingArgs() > 0 {
		if err := caddyutils.ExpectString(d, "to"); err != nil {
			return err
		}
		dest := server.MapDest{Weight: 100}
		if err := caddyutils.ParseString(d, &dest.Subject); err != nil {
			return err
		}
		mapping.MapDest = append(mapping.MapDest, &dest)
		account.Mappings = append(account.Mappings, &mapping)
	} else {
		// Long syntax
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			if d.Val() != "to" {
				return d.Errf("unrecognized subject mapping subdirective: %s", d.Val())
			}
			dest := server.MapDest{}
			if err := caddyutils.ParseString(d, &dest.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := caddyutils.ExpectString(d, "weight"); err != nil {
					return err
				}
				if err := caddyutils.ParseUInt8(d, &dest.Weight); err != nil {
					return err
				}
			} else {
				dest.Weight = 100
			}
			mapping.MapDest = append(mapping.MapDest, &dest)
		}
		account.Mappings = append(account.Mappings, &mapping)
	}
	return nil
}

// parseAccounts parses the "accounts" directive found in the Caddyfile "nats_server" option block.
func parseAccounts(d *caddyfile.Dispenser, accounts *[]*embedded.Account) error {
	if accounts == nil {
		return d.Err("internal error: accounts is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		acc := embedded.Account{Name: d.Val()}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "jetstream":
				if err := caddyutils.ParseBool(d, &acc.JetStream); err != nil {
					return err
				}
			case "map_subject":
				acc.Mappings = fnutils.DefaultIfEmpty(acc.Mappings, []*embedded.SubjectMapping{})
				if err := parseSubjectMapping(d, &acc); err != nil {
					return err
				}
			case "export_service":
				if acc.Services == nil {
					acc.Services = &embedded.Services{}
				}
				if acc.Services.Export == nil {
					acc.Services.Export = []embedded.ServiceExport{}
				}
				export := embedded.ServiceExport{}
				if err := caddyutils.ParseString(d, &export.Subject); err != nil {
					return err
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "to"); err != nil {
						return err
					}
					if err := caddyutils.ParseStringArray(d, &export.To, false); err != nil {
						return err
					}
				}
				acc.Services.Export = append(acc.Services.Export, export)
			case "import_service":
				if acc.Services == nil {
					acc.Services = &embedded.Services{}
				}
				if acc.Services.Import == nil {
					acc.Services.Import = []embedded.ServiceImport{}
				}
				import_ := embedded.ServiceImport{}
				if err := caddyutils.ParseString(d, &import_.Subject); err != nil {
					return err
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "from"); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &import_.Account); err != nil {
						return err
					}
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "to"); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &import_.To); err != nil {
						return err
					}
				}
				acc.Services.Import = append(acc.Services.Import, import_)
			case "export_stream":
				if acc.Streams == nil {
					acc.Streams = &embedded.Streams{}
				}
				if acc.Streams.Export == nil {
					acc.Streams.Export = []embedded.StreamExport{}
				}
				export := embedded.StreamExport{}
				if err := caddyutils.ParseString(d, &export.Subject); err != nil {
					return err
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "to"); err != nil {
						return err
					}
					if err := caddyutils.ParseStringArray(d, &export.To, false); err != nil {
						return err
					}
				}
				acc.Streams.Export = append(acc.Streams.Export, export)
			case "import_stream":
				if acc.Streams == nil {
					acc.Streams = &embedded.Streams{}
				}
				if acc.Streams.Import == nil {
					acc.Streams.Import = []embedded.StreamImport{}
				}
				import_ := embedded.StreamImport{}
				if err := caddyutils.ParseString(d, &import_.Subject); err != nil {
					return err
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "from"); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &import_.Account); err != nil {
						return err
					}
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ExpectString(d, "to"); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &import_.To); err != nil {
						return err
					}
				}
				acc.Streams.Import = append(acc.Streams.Import, import_)
			default:
				return d.Errf("unrecognized account subdirective: %s", d.Val())
			}
		}
		*accounts = append(*accounts, &acc)
	}
	return nil
}

// parseAuthUsers parses the "users" directive found in the Caddyfile "nats_server" option block.
func parseAuthUsers(d *caddyfile.Dispenser, auth *embedded.AuthorizationMap) error {
	if auth == nil {
		return d.Err("internal error: authorization map is nil. Please open a bug report.")
	}
	if auth.Users == nil {
		return d.Err("internal error: authorization map users field is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		user := embedded.User{User: name}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "password":
				if err := caddyutils.ParseString(d, &user.Password); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized user subdirective: %s", d.Val())
			}
		}
		auth.Users = append(auth.Users, user)
	}
	return nil
}

// parseJetStream parses the "jetstream" directive found in the Caddyfile "nats_server" option block.
func parseJetStream(d *caddyfile.Dispenser, jsopts *embedded.JetStream) error {
	// Make sure we have o JetStream config
	if jsopts == nil {
		return d.Err("internal error: jetstream config is nil. Please open a bug report.")
	}
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
				if err := caddyutils.ParseString(d, &jsopts.Domain); err != nil {
					return err
				}
			case "store", "store_dir", "store_directory":
				if err := caddyutils.ParseString(d, &jsopts.StoreDir); err != nil {
					return err
				}
			case "max_memory":
				if err := caddyutils.ParseByteSizeI64(d, &jsopts.MaxMemory); err != nil {
					return err
				}
			case "max_file", "max_disk":
				if err := caddyutils.ParseByteSizeI64(d, &jsopts.MaxFile); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized jetstream subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseMqtt parses the "mqtt" directive found in the Caddyfile "nats_server" option block.
func parseMqtt(d *caddyfile.Dispenser, mqttopts *embedded.MQTT) error {
	// Make sure we have o MQTT config
	if mqttopts == nil {
		return d.Err("internal error: mqtt config is nil. Please open a bug report.")
	}
	if d.NextArg() {
		// Short syntax
		port, err := parseutils.ParsePort(d.Val())
		if err != nil {
			return d.Errf("invalid mqtt port: %v", err)
		}
		mqttopts.Port = port
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("mqtt short syntax requires exactly one port number")
		}
	} else {
		// Long syntax
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "no_tls":
				if err := caddyutils.ParseBool(d, &mqttopts.NoTLS); err != nil {
					return err
				}
			case "host":
				if err := caddyutils.ParseString(d, &mqttopts.Host); err != nil {
					return err
				}
			case "port":
				if err := caddyutils.ParsePort(d, &mqttopts.Port); err != nil {
					return err
				}
			case "jetstream_domain":
				if err := caddyutils.ParseString(d, &mqttopts.JSDomain); err != nil {
					return err
				}
			case "stream_replicas":
				if err := caddyutils.ParseInt(d, &mqttopts.StreamReplicas); err != nil {
					return err
				}
			case "user", "username":
				if err := caddyutils.ParseString(d, &mqttopts.Username); err != nil {
					return err
				}
			case "password":
				if err := caddyutils.ParseString(d, &mqttopts.Password); err != nil {
					return err
				}
			case "no_auth_user":
				if err := caddyutils.ParseString(d, &mqttopts.NoAuthUser); err != nil {
					return err
				}
			case "tls":
				mqttopts.TLS = fnutils.DefaultIfNil(mqttopts.TLS, &embedded.TLSMap{})
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

// parseWebsocket parses the "websocket" directive found in the Caddyfile "nats_server" option block.
func parseWebsocket(d *caddyfile.Dispenser, wsopts *embedded.Websocket) error {
	// Make sure we have o Websocket config
	if wsopts == nil {
		return d.Err("internal error: websocket config is nil. Please open a bug report.")
	}
	if d.NextArg() {
		port, err := parseutils.ParsePort(d.Val())
		if err != nil {
			return d.Errf("invalid websocket port: %v", err)
		}
		wsopts.Port = port
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("websocket short syntax requires exactly one port number")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "no_tls":
				if err := caddyutils.ParseBool(d, &wsopts.NoTLS); err != nil {
					return err
				}
			case "host":
				if err := caddyutils.ParseString(d, &wsopts.Host); err != nil {
					return err
				}
			case "port":
				if err := caddyutils.ParsePort(d, &wsopts.Port); err != nil {
					return err
				}
			case "advertise", "client_advertise":
				if err := caddyutils.ParseString(d, &wsopts.Advertise); err != nil {
					return err
				}
			case "user", "username":
				if err := caddyutils.ParseString(d, &wsopts.Username); err != nil {
					return err
				}
			case "password":
				if err := caddyutils.ParseString(d, &wsopts.Password); err != nil {
					return err
				}
			case "no_auth_user":
				if err := caddyutils.ParseString(d, &wsopts.NoAuthUser); err != nil {
					return err
				}
			case "compression", "enable_compression":
				if err := caddyutils.ParseBool(d, &wsopts.Compression); err != nil {
					return err
				}
			case "same_origin", "require_same_origin":
				if err := caddyutils.ParseBool(d, &wsopts.SameOrigin); err != nil {
					return err
				}
			case "allowed_origins":
				if err := caddyutils.ParseStringArray(d, &wsopts.AllowedOrigins, false); err != nil {
					return err
				}
			case "jwt_cookie":
				if err := caddyutils.ParseString(d, &wsopts.JWTCookie); err != nil {
					return err
				}
			case "tls":
				wsopts.TLS = fnutils.DefaultIfNil(wsopts.TLS, &embedded.TLSMap{})
				if err := parseTLS(d, wsopts.TLS); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized websocket subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseLeafnodes parse the "leafnodes" directive found in the Caddyfile "nats_server" option block.
func parseLeafnodes(d *caddyfile.Dispenser, leafopts *embedded.Leafnode) error {
	// Make sure we have o LeafNode config
	if leafopts == nil {
		return d.Err("internal error: leafnode config is nil. Please open a bug report.")
	}
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
				if err := caddyutils.ParseString(d, &leafopts.Host); err != nil {
					return err
				}
			case "port":
				if err := caddyutils.ParsePort(d, &leafopts.Port); err != nil {
					return err
				}
			case "advertise":
				if err := caddyutils.ParseString(d, &leafopts.Advertise); err != nil {
					return err
				}
			case "no_tls":
				if err := caddyutils.ParseBool(d, &leafopts.NoTLS); err != nil {
					return err
				}
			case "tls":
				leafopts.TLS = fnutils.DefaultIfNil(leafopts.TLS, &embedded.TLSMap{})
				if err := parseTLS(d, leafopts.TLS); err != nil {
					return err
				}
			case "remotes":
				if err := parseRemoteLeafnodes(d, &leafopts.Remotes); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// parseRemoteLeafnodes parse the "remote_leafnodes" directive found in the Caddyfile "nats_server" option block.
func parseRemoteLeafnodes(d *caddyfile.Dispenser, remotes *[]embedded.Remote) error {
	if remotes == nil {
		return d.Err("internal error: remotes is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		remote := embedded.Remote{Url: d.Val()}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "url":
				if err := caddyutils.ParseString(d, &remote.Url); err != nil {
					return err
				}
			case "urls":
				if err := caddyutils.ParseStringArray(d, &remote.Urls, false); err != nil {
					return err
				}
			case "hub":
				if err := caddyutils.ParseBool(d, &remote.Hub); err != nil {
					return err
				}
			case "deny_import":
				if err := caddyutils.ParseStringArray(d, &remote.DenyImports, false); err != nil {
					return err
				}
			case "deny_export":
				if err := caddyutils.ParseStringArray(d, &remote.DenyExports, false); err != nil {
					return err
				}
			case "account":
				if err := caddyutils.ParseString(d, &remote.Account); err != nil {
					return err
				}
			case "credentials":
				if err := caddyutils.ParseString(d, &remote.Credentials); err != nil {
					return err
				}
			case "websocket":
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "compression":
						if err := caddyutils.ParseBool(d, &remote.Websocket.Compression); err != nil {
							return err
						}
					case "no_masking":
						if err := caddyutils.ParseBool(d, &remote.Websocket.NoMasking); err != nil {
							return err
						}
					default:
						return d.Errf("unrecognized remote leafnode websocket subdirective: %s", d.Val())
					}
				}
			default:
				return d.Errf("unrecognized remote leafnode subdirective: %s", d.Val())
			}
		}
		*remotes = append(*remotes, remote)
	}
	return nil
}

// parseCacheResolver parses the "cache_resolver" directive found in the Caddyfile "nats_server" option block.
func parseCacheResolver(d *caddyfile.Dispenser, resolveropts *embedded.CacheAccountResolver) error {
	// Make sure we have o CacheAccountResolver config
	if resolveropts == nil {
		return d.Err("internal error: cache resolver config is nil. Please open a bug report.")
	}
	// Short syntax
	if d.NextArg() {
		resolveropts.Path = d.Val()
		if d.NextArg() {
			return d.Err("cache resolver short syntax requires exactly one path")
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("cache resolver short syntax requires exactly one path")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "path":
				if err := caddyutils.ParseString(d, &resolveropts.Path); err != nil {
					return err
				}
			case "limit":
				if err := caddyutils.ParseInt(d, &resolveropts.Limit); err != nil {
					return err
				}
			case "ttl":
				if err := caddyutils.ParseDuration(d, &resolveropts.TTL); err != nil {
					return err
				}
			case "preload":
				if err := caddyutils.ParseStringArray(d, &resolveropts.Preload, false); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized cache resolver subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseFullResolver parses the "full_resolver" directive found in the Caddyfile "nats_server" option block.
func parseFullResolver(d *caddyfile.Dispenser, resolveropts *embedded.FullAccountResolver) error {
	// Make sure we have o FullAccountResolver config
	if resolveropts == nil {
		return d.Err("internal error: full resolver config is nil. Please open a bug report.")
	}
	// Short syntax
	if d.NextArg() {
		resolveropts.Path = d.Val()
		if d.NextArg() {
			return d.Err("full resolver short syntax requires exactly one path")
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("full resolver short syntax requires exactly one path")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "path":
				if err := caddyutils.ParseString(d, &resolveropts.Path); err != nil {
					return err
				}
			case "limit":
				if err := caddyutils.ParseInt64(d, &resolveropts.Limit); err != nil {
					return err
				}
			case "sync", "sync_interval":
				if err := caddyutils.ParseDuration(d, &resolveropts.SyncInterval); err != nil {
					return err
				}
			case "allow_delete":
				if err := caddyutils.ParseBool(d, &resolveropts.AllowDelete); err != nil {
					return err
				}
			case "hard_delete":
				if err := caddyutils.ParseBool(d, &resolveropts.HardDelete); err != nil {
					return err
				}
			case "preload":
				if err := caddyutils.ParseStringArray(d, &resolveropts.Preload, false); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized full resolver subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseMemoryResolver parses the "memory_resolver" directive found in the Caddyfile "nats_server" option block.
func parseMemoryResolver(d *caddyfile.Dispenser, resolveropts *embedded.MemoryAccountResolver) error {
	// Make sure we have o MemoryAccountResolver config
	if resolveropts == nil {
		return d.Err("internal error: memory resolver config is nil. Please open a bug report.")
	}
	if d.NextArg() {
		return d.Err("memory resolver does not take any argument")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "limit":
			if err := caddyutils.ParseInt(d, &resolveropts.Limit); err != nil {
				return err
			}
		case "preload":
			if err := caddyutils.ParseStringArray(d, &resolveropts.Preload, false); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized memory resolver subdirective: %s", d.Val())
		}
	}
	return nil
}

// parseTLS parses the "tls" directive found in the Caddyfile "nats_server" option block.
func parseTLS(d *caddyfile.Dispenser, tlsOpts *embedded.TLSMap) error {
	if tlsOpts == nil {
		return d.Err("internal error: tlsOpts is nil. Please open a bug report.")
	}
	caddyutils.ParseStringArray(d, &tlsOpts.Subjects, false)
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "subjects":
			if err := caddyutils.ParseStringArray(d, &tlsOpts.Subjects, false); err != nil {
				return err
			}
		case "allow_non_tls":
			if err := caddyutils.ParseBool(d, &tlsOpts.AllowNonTLS); err != nil {
				return err
			}
		case "cert_file":
			if err := caddyutils.ParseString(d, &tlsOpts.CertFile); err != nil {
				return err
			}
		case "cert_match":
			if err := caddyutils.ParseString(d, &tlsOpts.CertMatch); err != nil {
				return err
			}
		case "cert_match_by":
			if err := caddyutils.ParseString(d, &tlsOpts.CertMatchBy); err != nil {
				return err
			}
		case "key_file":
			if err := caddyutils.ParseString(d, &tlsOpts.KeyFile); err != nil {
				return err
			}
		case "ca_file":
			if err := caddyutils.ParseString(d, &tlsOpts.CaFile); err != nil {
				return err
			}
		case "verify":
			if err := caddyutils.ParseBool(d, &tlsOpts.Verify); err != nil {
				return err
			}
		case "insecure":
			if err := caddyutils.ParseBool(d, &tlsOpts.Insecure); err != nil {
				return err
			}
		case "map":
			if err := caddyutils.ParseBool(d, &tlsOpts.Map); err != nil {
				return err
			}
		case "check_known_urls":
			if err := caddyutils.ParseBool(d, &tlsOpts.CheckKnownURLs); err != nil {
				return err
			}
		case "rate_limit":
			if err := caddyutils.ParseInt64(d, &tlsOpts.RateLimit); err != nil {
				return err
			}
		case "ciphers":
			if err := caddyutils.ParseStringArray(d, &tlsOpts.Ciphers, false); err != nil {
				return err
			}
		case "curve_preferences":
			if err := caddyutils.ParseStringArray(d, &tlsOpts.CurvePreferences, false); err != nil {
				return err
			}
		case "pinned_certs":
			if err := caddyutils.ParseStringArray(d, &tlsOpts.PinnedCerts, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseMetrics parses the "metrics" directive found in the Caddyfile "nats_server" option block.
func parseMetrics(d *caddyfile.Dispenser, metricopts *embedded.Metrics) error {
	// Make sure we have o Metrics config
	if metricopts == nil {
		return d.Err("internal error: metrics config is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "server_label":
			if err := caddyutils.ParseString(d, &metricopts.ServerLabel); err != nil {
				return err
			}
		case "server_url":
			if err := caddyutils.ParseString(d, &metricopts.ServerUrl); err != nil {
				return err
			}
		case "healthz":
			if err := caddyutils.ParseBool(d, &metricopts.Healthz); err != nil {
				return err
			}
		case "connz":
			if err := caddyutils.ParseBool(d, &metricopts.Connz); err != nil {
				return err
			}
			for d.NextArg() {
				switch d.Val() {
				case "detailed":
					if err := caddyutils.ParseBool(d, &metricopts.ConnzDetailed); err != nil {
						return err
					}
				default:
					return d.Err("invalid metrics connz option")
				}
			}
		case "connz_detailed":
			if err := caddyutils.ParseBool(d, &metricopts.ConnzDetailed); err != nil {
				return err
			}
		case "subz":
			if err := caddyutils.ParseBool(d, &metricopts.Subz); err != nil {
				return err
			}
		case "routez":
			if err := caddyutils.ParseBool(d, &metricopts.Routez); err != nil {
				return err
			}
		case "gatewayz":
			if err := caddyutils.ParseBool(d, &metricopts.Gatewayz); err != nil {
				return err
			}
		case "leafz":
			if err := caddyutils.ParseBool(d, &metricopts.Leafz); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
