// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddynats

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
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
	d.Next() // skip "nats"
	server := &Server{}
	if err := server.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	a.Server = server
	return nil
}

func (s *Server) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	s.Options = &Options{}
	if err := s.Options.UnmarshalCaddyfile(d); err != nil {
		return err
	}
	return nil
}

func (o *Options) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// Do not expect any argument but o block instead
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "nats", "nats_server", "server":
			if err := o.UnmarshalCaddyfile(d); err != nil {
				return err
			}
		case "cert_issuer":
			module := ""
			if err := parser.ParseString(d, &module); err != nil {
				return err
			}
			unm, err := caddyfile.UnmarshalModule(d, "tls.issuance."+module)
			if err != nil {
				return err
			}
			if o.AutomationPolicyTemplate == nil {
				o.AutomationPolicyTemplate = &caddytls.AutomationPolicy{}
			}
			if o.AutomationPolicyTemplate.IssuersRaw == nil {
				o.AutomationPolicyTemplate.IssuersRaw = []json.RawMessage{}
			}
			o.AutomationPolicyTemplate.IssuersRaw = append(o.AutomationPolicyTemplate.IssuersRaw, caddyconfig.JSONModuleObject(unm, "module", module, nil))
		case "no_tls":
			if err := parser.ParseBool(d, &o.NoTLS); err != nil {
				return err
			}
		case "name", "server_name":
			if err := parser.ParseString(d, &o.ServerName); err != nil {
				return err
			}
		case "tags", "server_tags":
			// o.ServerTags = fnutils.DefaultIfEmptyMap(o.ServerTags, map[string]string{})
			if err := parser.ParseStringMap(d, &o.ServerTags, parser.Inline(parser.Separator(":"))); err != nil {
				return err
			}
		case "host", "server_host":
			if err := parser.ParseString(d, &o.Host); err != nil {
				return err
			}
		case "port", "server_port":
			if err := parser.ParseNetworkPort(d, &o.Port); err != nil {
				return err
			}
		case "advertise", "client_advertise":
			if err := parser.ParseString(d, &o.Advertise); err != nil {
				return err
			}
		case "debug", "enable_debug":
			if err := parser.ParseBool(d, &o.Debug); err != nil {
				return err
			}
		case "trace", "enable_trace":
			if err := parser.ParseBool(d, &o.Trace); err != nil {
				return err
			}
			if o.Trace {
				o.Debug = true
			}
		case "trace_verbose", "enable_trace_verbose":
			if err := parser.ParseBool(d, &o.TraceVerbose); err != nil {
				return err
			}
			if o.TraceVerbose {
				o.Debug = true
				o.Trace = true
			}
		case "http_port", "monitoring_port":
			if err := parser.ParseNetworkPort(d, &o.HTTPPort); err != nil {
				return err
			}
		case "http_host", "monitoring_host":
			if err := parser.ParseString(d, &o.HTTPHost); err != nil {
				return err
			}
		case "https_port", "monitoring_tls_port":
			if err := parser.ParseNetworkPort(d, &o.HTTPSPort); err != nil {
				return err
			}
		case "http_base_path", "monitoring_base_path":
			if err := parser.ParseString(d, &o.HTTPBasePath); err != nil {
				return err
			}
		case "no_log", "disable_logging":
			if err := parser.ParseBool(d, &o.NoLog); err != nil {
				return err
			}
		case "no_sublist_cache", "disable_sublist_cache":
			if err := parser.ParseBool(d, &o.NoSublistCache); err != nil {
				return err
			}
		case "max_conn", "max_connections":
			if err := parser.ParseInt(d, &o.MaxConn); err != nil {
				return err
			}
		case "max_payload":
			if err := parser.ParseInt32ByteSize(d, &o.MaxPayload); err != nil {
				return err
			}
		case "max_pending":
			if err := parser.ParseInt64ByteSize(d, &o.MaxPending); err != nil {
				return err
			}
		case "max_subs", "max_subscriptions":
			if err := parser.ParseInt(d, &o.MaxSubs); err != nil {
				return err
			}
		case "max_control_line":
			if err := parser.ParseInt32ByteSize(d, &o.MaxControlLine); err != nil {
				return err
			}
		case "ping_interval":
			if err := parser.ParseDuration(d, &o.PingInterval); err != nil {
				return err
			}
		case "max_pings_out", "ping_max":
			if err := parser.ParseInt(d, &o.MaxPingsOut); err != nil {
				return err
			}
		case "write_deadline":
			if err := parser.ParseDuration(d, &o.WriteDeadline); err != nil {
				return err
			}
		case "no_auth_user":
			if err := parser.ParseString(d, &o.NoAuthUser); err != nil {
				return err
			}
		case "system_account":
			if err := parser.ParseString(d, &o.SystemAccount); err != nil {
				return err
			}
		case "tls":
			o.TLS = fnutils.DefaultIfNil(o.TLS, &TLSMap{})
			if err := parseTLS(d, o.TLS); err != nil {
				return err
			}
		case "cluster":
			o.Cluster = fnutils.DefaultIfNil(o.Cluster, &Cluster{})
			if err := ParseCluster(d, o.Cluster); err != nil {
				return err
			}
		case "jetstream":
			o.JetStream = fnutils.DefaultIfNil(o.JetStream, &JetStream{})
			if err := ParseJetStream(d, o.JetStream); err != nil {
				return err
			}
		case "jetstream_max_disk", "jetstream_max_file":
			o.JetStream = fnutils.DefaultIfNil(o.JetStream, &JetStream{})
			if err := parser.ParseInt64ByteSize(d, &o.JetStream.MaxFile); err != nil {
				return err
			}
		case "jetstream_max_memory", "jetstream_max_mem":
			o.JetStream = fnutils.DefaultIfNil(o.JetStream, &JetStream{})
			if err := parser.ParseInt64ByteSize(d, &o.JetStream.MaxMemory); err != nil {
				return err
			}
		case "jetstream_domain":
			o.JetStream = fnutils.DefaultIfNil(o.JetStream, &JetStream{})
			if err := parser.ParseString(d, &o.JetStream.Domain); err != nil {
				return err
			}
		case "mqtt", "mqtt_server":
			o.Mqtt = fnutils.DefaultIfNil(o.Mqtt, &MQTT{})
			if err := ParseMqtt(d, o.Mqtt); err != nil {
				return err
			}
		case "websocket", "websocket_server":
			o.Websocket = fnutils.DefaultIfNil(o.Websocket, &Websocket{})
			if err := ParseWebsocket(d, o.Websocket); err != nil {
				return err
			}
		case "leafnodes", "leafnode", "leafnode_hub", "leafnode_remote", "leafnode_server":
			o.Leafnode = fnutils.DefaultIfNil(o.Leafnode, &Leafnode{})
			if err := ParseLeafnodes(d, o.Leafnode); err != nil {
				return err
			}
		case "operators", "operator":
			o.Operators = fnutils.DefaultIfEmpty(o.Operators, []string{})
			if err := parser.ParseStringArray(d, &o.Operators); err != nil {
				return err
			}
		case "auth", "authorization":
			o.Authorization = fnutils.DefaultIfNil(o.Authorization, &AuthorizationMap{})
			if err := ParseAuthorization(d, o.Authorization); err != nil {
				return err
			}
		case "auth_callout", "authorization_callout":
			o.AuthCallout = fnutils.DefaultIfNil(o.AuthCallout, &AuthCalloutMap{})
			if err := parseAuthCallout(d, o.AuthCallout); err != nil {
				return err
			}
		case "auth_account":
			o.AuthCallout = fnutils.DefaultIfNil(o.AuthCallout, &AuthCalloutMap{})
			if err := parser.ParseString(d, &o.AuthCallout.Account); err != nil {
				return err
			}
		case "account":
			o.Accounts = fnutils.DefaultIfEmpty(o.Accounts, []*Account{})
			acc := Account{}
			if err := parser.ParseString(d, &acc.Name); err != nil {
				return err
			}
			if err := ParseAccount(d, &acc); err != nil {
				return err
			}
			o.Accounts = append(o.Accounts, &acc)
		case "accounts":
			o.Accounts = fnutils.DefaultIfEmpty(o.Accounts, []*Account{})
			if err := ParseAccounts(d, &o.Accounts); err != nil {
				return err
			}
		case "users":
			o.Authorization = fnutils.DefaultIfNil(o.Authorization, &AuthorizationMap{})
			o.Authorization.Users = fnutils.DefaultIfEmpty(o.Authorization.Users, []User{})
			if err := parseAuthUsers(d, o.Authorization); err != nil {
				return err
			}
		case "metrics":
			o.Metrics = fnutils.DefaultIfNil(o.Metrics, &Metrics{})
			if err := ParseMetrics(d, o.Metrics); err != nil {
				return err
			}
		case "full_resolver":
			o.FullResolver = fnutils.DefaultIfNil(o.FullResolver, &FullAccountResolver{})
			if err := parseFullResolver(d, o.FullResolver); err != nil {
				return err
			}
		case "cache_resolver":
			o.CacheResolver = fnutils.DefaultIfNil(o.CacheResolver, &CacheAccountResolver{})
			if err := parseCacheResolver(d, o.CacheResolver); err != nil {
				return err
			}
		case "memory_resolver":
			o.MemoryResolver = fnutils.DefaultIfNil(o.MemoryResolver, &MemoryAccountResolver{})
			if err := parseMemoryResolver(d, o.MemoryResolver); err != nil {
				return err
			}
		case "resolver":
			return d.Err("resolver directive has been removed, use full_resolver, cache_resolver or memory_resolver instead")
		default:
			return d.Errf("unrecognized nats subdirective: %s", d.Val())
		}
	}
	return nil
}

func ParseCluster(d *caddyfile.Dispenser, cluster *Cluster) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "name":
			if err := parser.ParseString(d, &cluster.Name); err != nil {
				return err
			}
		case "host":
			if err := parser.ParseString(d, &cluster.Host); err != nil {
				return err
			}
		case "port":
			if err := parser.ParseNetworkPort(d, &cluster.Port); err != nil {
				return err
			}
		case "advertise":
			if err := parser.ParseString(d, &cluster.Advertise); err != nil {
				return err
			}
		case "routes", "route":
			if err := parser.ParseStringArray(d, &cluster.Routes); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseSubjectMapping(d *caddyfile.Dispenser, account *Account) error {
	if account == nil {
		return d.Err("internal error: account is nil. Please open a bug report.")
	}
	if account.Mappings == nil {
		return d.Err("internal error: mappings is nil. Please open a bug report.")
	}
	mapping := SubjectMapping{MapDest: []*server.MapDest{}}
	if err := parser.ParseString(d, &mapping.Subject); err != nil {
		return err
	}
	if d.CountRemainingArgs() > 0 {
		if err := parser.ExpectString(d, parser.Match("to")); err != nil {
			return err
		}
		dest := server.MapDest{Weight: 100}
		if err := parser.ParseString(d, &dest.Subject); err != nil {
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
			if err := parser.ParseString(d, &dest.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("weight")); err != nil {
					return err
				}
				if err := parser.ParseUint8(d, &dest.Weight); err != nil {
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

func ParseAccount(d *caddyfile.Dispenser, acc *Account) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "name":
			if err := parser.ParseString(d, &acc.Name); err != nil {
				return err
			}
		case "nkey":
			if err := parser.ParseString(d, &acc.NKey); err != nil {
				return err
			}
		case "jetstream":
			if err := parser.ParseBool(d, &acc.JetStream); err != nil {
				return err
			}
		case "map_subject":
			acc.Mappings = fnutils.DefaultIfEmpty(acc.Mappings, []*SubjectMapping{})
			if err := parseSubjectMapping(d, acc); err != nil {
				return err
			}
		case "export_service":
			if acc.Exports == nil {
				acc.Exports = &Exports{}
			}
			if acc.Exports.Services == nil {
				acc.Exports.Services = []ServiceExport{}
			}
			export := ServiceExport{}
			if err := parser.ParseString(d, &export.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("to")); err != nil {
					return err
				}
				if err := parser.ParseStringArray(d, &export.To); err != nil {
					return err
				}
			}
			acc.Exports.Services = append(acc.Exports.Services, export)
		case "import_service":
			if acc.Imports == nil {
				acc.Imports = &Imports{}
			}
			if acc.Imports.Services == nil {
				acc.Imports.Services = []ServiceImport{}
			}
			import_ := ServiceImport{}
			if err := parser.ParseString(d, &import_.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("from")); err != nil {
					return err
				}
				if err := parser.ParseString(d, &import_.Account); err != nil {
					return err
				}
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("to")); err != nil {
					return err
				}
				if err := parser.ParseString(d, &import_.To); err != nil {
					return err
				}
			}
			acc.Imports.Services = append(acc.Imports.Services, import_)
		case "export_stream":
			if acc.Exports == nil {
				acc.Exports = &Exports{}
			}
			if acc.Exports.Streams == nil {
				acc.Exports.Streams = []StreamExport{}
			}
			export := StreamExport{}
			if err := parser.ParseString(d, &export.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("to")); err != nil {
					return err
				}
				if err := parser.ParseStringArray(d, &export.To); err != nil {
					return err
				}
			}
			acc.Exports.Streams = append(acc.Exports.Streams, export)
		case "import_stream":
			if acc.Imports == nil {
				acc.Imports = &Imports{}
			}
			if acc.Imports.Streams == nil {
				acc.Imports.Streams = []StreamImport{}
			}
			import_ := StreamImport{}
			if err := parser.ParseString(d, &import_.Subject); err != nil {
				return err
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("from")); err != nil {
					return err
				}
				if err := parser.ParseString(d, &import_.Account); err != nil {
					return err
				}
			}
			if d.CountRemainingArgs() > 0 {
				if err := parser.ExpectString(d, parser.Match("to")); err != nil {
					return err
				}
				if err := parser.ParseString(d, &import_.To); err != nil {
					return err
				}
			}
			acc.Imports.Streams = append(acc.Imports.Streams, import_)
		case "auth_policy", "authorize":
			if acc.AuthorizationPolicies == nil {
				acc.AuthorizationPolicies = []*natsauth.AuthorizationPolicy{}
			}
			policy := natsauth.AuthorizationPolicy{}
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "match":
					if policy.MatchersRaw == nil {
						policy.MatchersRaw = map[string]json.RawMessage{}
					}
					if d.CountRemainingArgs() > 0 {
						module := ""
						if err := parser.ParseString(d, &module); err != nil {
							return err
						}
						unm, err := caddyfile.UnmarshalModule(d, "nats.matchers."+module)
						if err != nil {
							return err
						}
						matcher, ok := unm.(natsauth.AuthorizationMatcher)
						if !ok {
							return errors.New("matcher module is not a matcher")
						}
						policy.MatchersRaw[module] = caddyconfig.JSON(matcher, nil)
					} else {
						for nesting := d.Nesting(); d.NextBlock(nesting); {
							module := d.Val()
							unm, err := caddyfile.UnmarshalModule(d, "nats.matchers."+module)
							if err != nil {
								return err
							}
							matcher, ok := unm.(natsauth.AuthorizationMatcher)
							if !ok {
								return errors.New("matcher module is not a matcher")
							}
							policy.MatchersRaw[module] = caddyconfig.JSON(matcher, nil)
						}
					}

				case "callout":
					module := ""
					if err := parser.ParseString(d, &module); err != nil {
						return err
					}
					unm, err := caddyfile.UnmarshalModule(d, "nats.callouts."+module)
					if err != nil {
						return err
					}
					callout, ok := unm.(natsauth.AuthorizationCallout)
					if !ok {
						return errors.New("callout module is not a callout")
					}
					policy.CalloutRaw = caddyconfig.JSONModuleObject(callout, "module", module, nil)
				default:
					return d.Errf("unrecognized auth_policy subdirective: %s", d.Val())
				}
			}
			acc.AuthorizationPolicies = append(acc.AuthorizationPolicies, &policy)
		case "connect_leaf", "leafnode_connection", "leafnode_connect":
			remote := &Remote{
				Account: acc.Name,
			}
			if err := ParseRemoteLeafnode(d, remote); err != nil {
				return err
			}
			if acc.LeafnodeConnections == nil {
				acc.LeafnodeConnections = []*Remote{}
			}
			acc.LeafnodeConnections = append(acc.LeafnodeConnections, remote)
		case "stream":
			acc.JetStream = true
			acc.Streams = fnutils.DefaultIfEmpty(acc.Streams, []*natsclient.Stream{})
			stream := natsclient.Stream{}
			if err := stream.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			acc.Streams = append(acc.Streams, &stream)
		case "flow", "data_flow":
			acc.JetStream = true
			acc.Flows = fnutils.DefaultIfEmpty(acc.Flows, []*Flow{})
			flow := Flow{}
			if err := flow.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			acc.Flows = append(acc.Flows, &flow)
		case "service":
			acc.Services = fnutils.DefaultIfEmpty(acc.Services, []json.RawMessage{})
			provider, err := natsclient.LoadRawServiceProvider(d, "type")
			if err != nil {
				return err
			}
			acc.Services = append(acc.Services, provider)
		case "object_store":
			acc.ObjectStores = fnutils.DefaultIfEmpty(acc.ObjectStores, []*natsclient.ObjectStore{})
			store := natsclient.ObjectStore{}
			if err := store.UnmarshalCaddyfile(d); err != nil {
				return err
			}
			acc.ObjectStores = append(acc.ObjectStores, &store)
		default:
			return d.Errf("unrecognized account subdirective: %s", d.Val())
		}
	}
	return nil
}

// ParseAccounts parses the "accounts" directive found in the Caddyfile "nats" option block.
func ParseAccounts(d *caddyfile.Dispenser, accounts *[]*Account) error {
	if accounts == nil {
		return d.Err("internal error: accounts is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		acc := Account{Name: d.Val()}
		if err := ParseAccount(d, &acc); err != nil {
			return err
		}
		*accounts = append(*accounts, &acc)
	}
	return nil
}

func ParseAuthorization(d *caddyfile.Dispenser, auth *AuthorizationMap) error {
	if auth == nil {
		return d.Err("internal error: authorization map is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "token":
			if err := parser.ParseString(d, &auth.Token); err != nil {
				return err
			}
		case "user":
			if err := parser.ParseString(d, &auth.User); err != nil {
				return err
			}
		case "password":
			if err := parser.ParseString(d, &auth.Password); err != nil {
				return err
			}
		case "timeout":
			if err := parser.ParseDuration(d, &auth.Timeout); err != nil {
				return err
			}
		case "users":
			auth.Users = fnutils.DefaultIfEmpty(auth.Users, []User{})
			if err := parseAuthUsers(d, auth); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized authorization subdirective: %s", d.Val())
		}
	}
	return nil
}

func parseAuthCallout(d *caddyfile.Dispenser, dest *AuthCalloutMap) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "issuer":
			if err := parser.ParseString(d, &dest.Issuer); err != nil {
				return err
			}
		case "account":
			if err := parser.ParseString(d, &dest.Account); err != nil {
				return err
			}
		case "auth_users":
			if err := parser.ParseStringArray(d, &dest.AuthUsers); err != nil {
				return err
			}
		case "xkey":
			if err := parser.ParseString(d, &dest.XKey); err != nil {
				return err
			}
		case "signing_key":
			if err := parser.ParseString(d, &dest.SigningKey); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized auth_callout subdirective: %s", d.Val())
		}
	}
	return nil
}

// parseAuthUsers parses the "users" directive found in the Caddyfile "nats" option block.
func parseAuthUsers(d *caddyfile.Dispenser, auth *AuthorizationMap) error {
	if auth == nil {
		return d.Err("internal error: authorization map is nil. Please open a bug report.")
	}
	if auth.Users == nil {
		return d.Err("internal error: authorization map users field is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		name := d.Val()
		user := User{User: name}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "password":
				if err := parser.ParseString(d, &user.Password); err != nil {
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

// parseJetStream parses the "jetstream" directive found in the Caddyfile "nats" option block.
func ParseJetStream(d *caddyfile.Dispenser, jsopts *JetStream) error {
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
				if err := parser.ParseString(d, &jsopts.Domain); err != nil {
					return err
				}
			case "store", "store_dir", "store_directory":
				if err := parser.ParseString(d, &jsopts.StoreDir); err != nil {
					return err
				}
			case "max_memory":
				if err := parser.ParseInt64ByteSize(d, &jsopts.MaxMemory); err != nil {
					return err
				}
			case "max_file", "max_disk":
				if err := parser.ParseInt64ByteSize(d, &jsopts.MaxFile); err != nil {
					return err
				}
			case "unique_tag":
				if err := parser.ParseString(d, &jsopts.UniqueTag); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized jetstream subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// ParseMqtt parses the "mqtt" directive found in the Caddyfile "nats" option block.
func ParseMqtt(d *caddyfile.Dispenser, mqttopts *MQTT) error {
	// Make sure we have o MQTT config
	if mqttopts == nil {
		return d.Err("internal error: mqtt config is nil. Please open a bug report.")
	}
	if d.NextArg() {
		// Short syntax
		if err := parser.ParseNetworkPort(d, &mqttopts.Port, parser.Inplace()); err != nil {
			return d.Errf("invalid mqtt port: %v", err)
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("mqtt short syntax requires exactly one port number")
		}
	} else {
		// Long syntax
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "no_tls":
				if err := parser.ParseBool(d, &mqttopts.NoTLS); err != nil {
					return err
				}
			case "host":
				if err := parser.ParseString(d, &mqttopts.Host); err != nil {
					return err
				}
			case "port":
				if err := parser.ParseNetworkPort(d, &mqttopts.Port); err != nil {
					return err
				}
			case "jetstream_domain":
				if err := parser.ParseString(d, &mqttopts.JSDomain); err != nil {
					return err
				}
			case "stream_replicas":
				if err := parser.ParseInt(d, &mqttopts.StreamReplicas); err != nil {
					return err
				}
			case "user", "username":
				if err := parser.ParseString(d, &mqttopts.Username); err != nil {
					return err
				}
			case "password":
				if err := parser.ParseString(d, &mqttopts.Password); err != nil {
					return err
				}
			case "no_auth_user":
				if err := parser.ParseString(d, &mqttopts.NoAuthUser); err != nil {
					return err
				}
			case "tls":
				mqttopts.TLS = fnutils.DefaultIfNil(mqttopts.TLS, &TLSMap{})
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

// ParseWebsocket parses the "websocket" directive found in the Caddyfile "nats" option block.
func ParseWebsocket(d *caddyfile.Dispenser, wsopts *Websocket) error {
	// Make sure we have o Websocket config
	if wsopts == nil {
		return d.Err("internal error: websocket config is nil. Please open a bug report.")
	}
	if d.NextArg() {
		// Short syntax
		if err := parser.ParseNetworkPort(d, &wsopts.Port, parser.Inplace()); err != nil {
			return d.Errf("invalid websocket port: %v", err)
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("websocket short syntax requires exactly one port number")
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "no_tls":
				if err := parser.ParseBool(d, &wsopts.NoTLS); err != nil {
					return err
				}
			case "host":
				if err := parser.ParseString(d, &wsopts.Host); err != nil {
					return err
				}
			case "port":
				if err := parser.ParseNetworkPort(d, &wsopts.Port); err != nil {
					return err
				}
			case "advertise", "client_advertise":
				if err := parser.ParseString(d, &wsopts.Advertise); err != nil {
					return err
				}
			case "user", "username":
				if err := parser.ParseString(d, &wsopts.Username); err != nil {
					return err
				}
			case "password":
				if err := parser.ParseString(d, &wsopts.Password); err != nil {
					return err
				}
			case "no_auth_user":
				if err := parser.ParseString(d, &wsopts.NoAuthUser); err != nil {
					return err
				}
			case "compression", "enable_compression":
				if err := parser.ParseBool(d, &wsopts.Compression); err != nil {
					return err
				}
			case "same_origin", "require_same_origin":
				if err := parser.ParseBool(d, &wsopts.SameOrigin); err != nil {
					return err
				}
			case "allowed_origins":
				if err := parser.ParseStringArray(d, &wsopts.AllowedOrigins); err != nil {
					return err
				}
			case "jwt_cookie":
				if err := parser.ParseString(d, &wsopts.JWTCookie); err != nil {
					return err
				}
			case "tls":
				wsopts.TLS = fnutils.DefaultIfNil(wsopts.TLS, &TLSMap{})
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

// ParseLeafnodes parse the "leafnodes" directive found in the Caddyfile "nats" option block.
func ParseLeafnodes(d *caddyfile.Dispenser, leafopts *Leafnode) error {
	// Make sure we have o LeafNode config
	if leafopts == nil {
		return d.Err("internal error: leafnode config is nil. Please open a bug report.")
	}
	// Short syntax
	if d.NextArg() {
		if err := parser.ParseNetworkPort(d, &leafopts.Port, parser.Inplace()); err != nil {
			return d.Errf("invalid leafnodes port: %v", err)
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Err("leafnodes short syntax requires exactly one port number")
		}
		// Long syntax
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "host":
				if err := parser.ParseString(d, &leafopts.Host); err != nil {
					return err
				}
			case "port":
				if err := parser.ParseNetworkPort(d, &leafopts.Port); err != nil {
					return err
				}
			case "advertise":
				if err := parser.ParseString(d, &leafopts.Advertise); err != nil {
					return err
				}
			case "no_tls":
				if err := parser.ParseBool(d, &leafopts.NoTLS); err != nil {
					return err
				}
			case "tls":
				leafopts.TLS = fnutils.DefaultIfNil(leafopts.TLS, &TLSMap{})
				if err := parseTLS(d, leafopts.TLS); err != nil {
					return err
				}
			case "remotes":
				if err := ParseRemoteLeafnodes(d, &leafopts.Remotes); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ParseRemoteLeafnodes parse the "remote_leafnodes" directive found in the Caddyfile "nats" option block.
func ParseRemoteLeafnodes(d *caddyfile.Dispenser, remotes *[]*Remote) error {
	if remotes == nil {
		return d.Err("internal error: remotes is nil. Please open a bug report.")
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		remote := Remote{Urls: []string{d.Val()}}
		if err := ParseRemoteLeafnode(d, &remote); err != nil {
			return err
		}
		*remotes = append(*remotes, &remote)
	}
	return nil
}

func ParseRemoteLeafnode(d *caddyfile.Dispenser, remote *Remote) error {
	if d.CountRemainingArgs() > 0 {
		if err := parser.ParseStringArray(d, &remote.Urls); err != nil {
			return err
		}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "urls":
			if err := parser.ParseStringArray(d, &remote.Urls); err != nil {
				return err
			}
		case "hub":
			if err := parser.ParseBool(d, &remote.Hub); err != nil {
				return err
			}
		case "deny_import":
			if err := parser.ParseStringArray(d, &remote.DenyImports); err != nil {
				return err
			}
		case "deny_export":
			if err := parser.ParseStringArray(d, &remote.DenyExports); err != nil {
				return err
			}
		case "account":
			if err := parser.ParseString(d, &remote.Account); err != nil {
				return err
			}
		case "credentials":
			if err := parser.ParseString(d, &remote.Credentials); err != nil {
				return err
			}
		case "websocket":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "compression":
					if err := parser.ParseBool(d, &remote.Websocket.Compression); err != nil {
						return err
					}
				case "no_masking":
					if err := parser.ParseBool(d, &remote.Websocket.NoMasking); err != nil {
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
	return nil
}

// parseCacheResolver parses the "cache_resolver" directive found in the Caddyfile "nats" option block.
func parseCacheResolver(d *caddyfile.Dispenser, resolveropts *CacheAccountResolver) error {
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
				if err := parser.ParseString(d, &resolveropts.Path); err != nil {
					return err
				}
			case "limit":
				if err := parser.ParseInt(d, &resolveropts.Limit); err != nil {
					return err
				}
			case "ttl":
				if err := parser.ParseDuration(d, &resolveropts.TTL); err != nil {
					return err
				}
			case "preload":
				if err := parser.ParseStringArray(d, &resolveropts.Preload); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized cache resolver subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseFullResolver parses the "full_resolver" directive found in the Caddyfile "nats" option block.
func parseFullResolver(d *caddyfile.Dispenser, resolveropts *FullAccountResolver) error {
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
				if err := parser.ParseString(d, &resolveropts.Path); err != nil {
					return err
				}
			case "limit":
				if err := parser.ParseInt64(d, &resolveropts.Limit); err != nil {
					return err
				}
			case "sync", "sync_interval":
				if err := parser.ParseDuration(d, &resolveropts.SyncInterval); err != nil {
					return err
				}
			case "allow_delete":
				if err := parser.ParseBool(d, &resolveropts.AllowDelete); err != nil {
					return err
				}
			case "hard_delete":
				if err := parser.ParseBool(d, &resolveropts.HardDelete); err != nil {
					return err
				}
			case "preload":
				if err := parser.ParseStringArray(d, &resolveropts.Preload); err != nil {
					return err
				}
			default:
				return d.Errf("unrecognized full resolver subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// parseMemoryResolver parses the "memory_resolver" directive found in the Caddyfile "nats" option block.
func parseMemoryResolver(d *caddyfile.Dispenser, resolveropts *MemoryAccountResolver) error {
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
			if err := parser.ParseInt(d, &resolveropts.Limit); err != nil {
				return err
			}
		case "preload":
			if err := parser.ParseStringArray(d, &resolveropts.Preload); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized memory resolver subdirective: %s", d.Val())
		}
	}
	return nil
}

// parseTLS parses the "tls" directive found in the Caddyfile "nats" option block.
func parseTLS(d *caddyfile.Dispenser, tlsOpts *TLSMap) error {
	if tlsOpts == nil {
		return d.Err("internal error: tlsOpts is nil. Please open a bug report.")
	}
	parser.ParseStringArray(d, &tlsOpts.Subjects, parser.AllowEmpty())
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "subjects":
			if err := parser.ParseStringArray(d, &tlsOpts.Subjects); err != nil {
				return err
			}
		case "allow_non_tls":
			if err := parser.ParseBool(d, &tlsOpts.AllowNonTLS); err != nil {
				return err
			}
		case "cert_file":
			if err := parser.ParseString(d, &tlsOpts.CertFile); err != nil {
				return err
			}
		case "cert_match":
			if err := parser.ParseString(d, &tlsOpts.CertMatch); err != nil {
				return err
			}
		case "cert_match_by":
			if err := parser.ParseString(d, &tlsOpts.CertMatchBy); err != nil {
				return err
			}
		case "key_file":
			if err := parser.ParseString(d, &tlsOpts.KeyFile); err != nil {
				return err
			}
		case "ca_file":
			if err := parser.ParseString(d, &tlsOpts.CaFile); err != nil {
				return err
			}
		case "verify":
			if err := parser.ParseBool(d, &tlsOpts.Verify); err != nil {
				return err
			}
		case "insecure":
			if err := parser.ParseBool(d, &tlsOpts.Insecure); err != nil {
				return err
			}
		case "map":
			if err := parser.ParseBool(d, &tlsOpts.Map); err != nil {
				return err
			}
		case "check_known_urls":
			if err := parser.ParseBool(d, &tlsOpts.CheckKnownURLs); err != nil {
				return err
			}
		case "rate_limit":
			if err := parser.ParseInt64(d, &tlsOpts.RateLimit); err != nil {
				return err
			}
		case "ciphers":
			if err := parser.ParseStringArray(d, &tlsOpts.Ciphers); err != nil {
				return err
			}
		case "curve_preferences":
			if err := parser.ParseStringArray(d, &tlsOpts.CurvePreferences); err != nil {
				return err
			}
		case "pinned_certs":
			if err := parser.ParseStringArray(d, &tlsOpts.PinnedCerts); err != nil {
				return err
			}
		}
	}
	return nil
}

// ParseMetrics parses the "metrics" directive found in the Caddyfile "nats" option block.
func ParseMetrics(d *caddyfile.Dispenser, metricopts *Metrics) error {
	// Make sure we have o Metrics config
	if metricopts == nil {
		return d.Err("internal error: metrics config is nil. Please open a bug report.")
	}
	if d.CountRemainingArgs() > 0 {
		for d.NextArg() {
			parseMetricInlineOption(d, metricopts)
		}
	} else {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			parseMetricOption(d, metricopts)
		}
	}
	return nil
}

func parseMetricInlineOption(d *caddyfile.Dispenser, metricopts *Metrics) error {
	switch d.Val() {
	case "healthz":
		metricopts.Healthz = true
	case "connz":
		metricopts.Connz = true
	case "connz_detailed":
		metricopts.ConnzDetailed = true
	case "subz":
		metricopts.Subz = true
	case "routez":
		metricopts.Routez = true
	case "gatewayz":
		metricopts.Gatewayz = true
	case "leafz":
		metricopts.Leafz = true
	case "all":
		metricopts.Healthz = true
		metricopts.Connz = true
		metricopts.ConnzDetailed = true
		metricopts.Subz = true
		metricopts.Routez = true
		metricopts.Gatewayz = true
		metricopts.Leafz = true
	default:
		return d.Errf("unrecognized inline metric option: %s", d.Val())
	}
	return nil
}

func parseMetricOption(d *caddyfile.Dispenser, metricopts *Metrics) error {
	switch d.Val() {
	case "server_label":
		if err := parser.ParseString(d, &metricopts.ServerLabel); err != nil {
			return err
		}
	case "server_url":
		if err := parser.ParseString(d, &metricopts.ServerUrl); err != nil {
			return err
		}
	case "healthz":
		if err := parser.ParseBool(d, &metricopts.Healthz); err != nil {
			return err
		}
	case "connz":
		if err := parser.ParseBool(d, &metricopts.Connz); err != nil {
			return err
		}
	case "connz_detailed":
		if err := parser.ParseBool(d, &metricopts.ConnzDetailed); err != nil {
			return err
		}
	case "subz":
		if err := parser.ParseBool(d, &metricopts.Subz); err != nil {
			return err
		}
	case "routez":
		if err := parser.ParseBool(d, &metricopts.Routez); err != nil {
			return err
		}
	case "gatewayz":
		if err := parser.ParseBool(d, &metricopts.Gatewayz); err != nil {
			return err
		}
	case "leafz":
		if err := parser.ParseBool(d, &metricopts.Leafz); err != nil {
			return err
		}
	default:
		return d.Errf("unrecognized subdirective: %s", d.Val())
	}
	return nil
}
