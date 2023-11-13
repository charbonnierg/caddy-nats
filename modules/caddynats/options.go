// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package caddynats

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/server/certstore"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/modules/caddynats/natsauth"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/modules/caddynats/natslogger"
	"github.com/quara-dev/beyond/modules/caddynats/natsmetrics"
	"github.com/quara-dev/beyond/modules/caddynats/natstls"
	"go.uber.org/zap"
)

type KeyStore interface {
	Provision(ctx caddy.Context) error
	GetKey(account string) (string, error)
}

// Options is the configuration for the NATS server
// It can be used to generate a server.Options struct.
// Default values will be used for missing fields.
type Options struct {
	// Private fields are generated when Provision() is called
	ctx              caddy.Context
	automateSubjects []string
	tlsApp           *caddytls.TLS
	logger           *zap.Logger
	systemAccount    *jwt.AccountClaims
	collectorOpts    *natsmetrics.CollectorOptions
	serverOpts       *server.Options
	authService      *authService

	ServerName               string                     `json:"name,omitempty"`
	ServerTags               map[string]string          `json:"tags,omitempty"`
	Host                     string                     `json:"host,omitempty"`
	Port                     int                        `json:"port,omitempty"`
	Advertise                string                     `json:"advertise,omitempty"`
	Debug                    bool                       `json:"debug,omitempty"`
	Trace                    bool                       `json:"trace,omitempty"`
	TraceVerbose             bool                       `json:"trace_verbose,omitempty"`
	HTTPHost                 string                     `json:"http_host,omitempty"`
	HTTPPort                 int                        `json:"http_port,omitempty"`
	HTTPSPort                int                        `json:"https_port,omitempty"`
	HTTPBasePath             string                     `json:"http_base_path,omitempty"`
	NoLog                    bool                       `json:"no_log,omitempty"`
	NoTLS                    bool                       `json:"no_tls,omitempty"`
	AutomationPolicyTemplate *caddytls.AutomationPolicy `json:"automation_policy,omitempty"`
	TLS                      *TLSMap                    `json:"tls,omitempty"`
	NoSublistCache           bool                       `json:"disable_sublist_cache,omitempty"`
	MaxConn                  int                        `json:"max_connections,omitempty"`
	MaxPayload               int32                      `json:"max_payload,omitempty"`
	MaxPending               int64                      `json:"max_pending,omitempty"`
	MaxClosedClients         int                        `json:"max_closed_clients,omitempty"`
	MaxSubs                  int                        `json:"max_subscriptions,omitempty"`
	MaxSubsTokens            uint8                      `json:"max_subscriptions_tokens,omitempty"`
	MaxControlLine           int32                      `json:"max_control_line,omitempty"`
	MaxTracedMsgLen          int                        `json:"max_traced_msg_len,omitempty"`
	MaxPingsOut              int                        `json:"max_pings_out,omitempty"`
	PingInterval             time.Duration              `json:"ping_interval,omitempty"`
	WriteDeadline            time.Duration              `json:"write_deadline,omitempty"`
	NoAuthUser               string                     `json:"no_auth_user,omitempty"`
	Operators                []string                   `json:"operators,omitempty"`
	SystemAccount            string                     `json:"system_account,omitempty"`
	Accounts                 []*Account                 `json:"accounts,omitempty"`
	AuthCallout              *AuthCalloutMap            `json:"auth_callout,omitempty"`
	Authorization            *AuthorizationMap          `json:"authorization,omitempty"`
	FullResolver             *FullAccountResolver       `json:"full_resolver,omitempty"`
	CacheResolver            *CacheAccountResolver      `json:"cache_resolver,omitempty"`
	MemoryResolver           *MemoryAccountResolver     `json:"memory_resolver,omitempty"`
	Cluster                  *Cluster                   `json:"cluster,omitempty"`
	Websocket                *Websocket                 `json:"websocket,omitempty"`
	Mqtt                     *MQTT                      `json:"mqtt,omitempty"`
	JetStream                *JetStream                 `json:"jetstream,omitempty"`
	Leafnode                 *Leafnode                  `json:"leafnode,omitempty"`
	Metrics                  *Metrics                   `json:"metrics,omitempty"`
}

// GetServerOptions returns the options for the NATS server.
func (o *Options) Provision(ctx caddy.Context) error {
	o.ctx = ctx
	o.automateSubjects = []string{}
	o.logger = ctx.Logger().Named("nats")
	if o.AutomationPolicyTemplate == nil {
		o.AutomationPolicyTemplate = &caddytls.AutomationPolicy{}
	}
	// Initialize options
	// DisableJetStreamBanner is always set to true
	// NoSigs is always set to true
	serverOpts := server.Options{
		DisableJetStreamBanner: true,
		NoSigs:                 true,
	}
	// Verify and set global options
	if err := o.setGlobalOpts(&serverOpts); err != nil {
		return err
	}
	// Static token auth
	if err := o.setTokenAuth(&serverOpts); err != nil {
		return err
	}
	// Static user/password auth
	if err := o.setUserPasswordAuth(&serverOpts); err != nil {
		return err
	}
	// Static multi-users auth
	if err := o.setUsersAuth(&serverOpts); err != nil {
		return err
	}
	// Static accounts auth
	if err := o.setAccountsAuth(&serverOpts); err != nil {
		return err
	}
	// Verify and set auth callout
	if err := o.setAuthCallout(&serverOpts); err != nil {
		return err
	}
	// Verify and set no auth user
	if err := o.setNoAuthUser(&serverOpts); err != nil {
		return err
	}
	// Verify and set monitoring options
	if err := o.setMonitoringOpts(&serverOpts); err != nil {
		return err
	}
	// Veirfy and set cluster options
	if err := o.setClusterOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set jetstream options
	if err := o.setJetStreamOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set leafnode options
	if err := o.setLeafnodeOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set websocket options
	if err := o.setWebsocketOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set mqtt options
	if err := o.setMqttOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set system account
	if err := o.setSystemAccountOpt(&serverOpts); err != nil {
		return err
	}
	// Verify and set operator mode options
	if err := o.setOperatorModeOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set resolver options
	if err := o.setResolverOpts(&serverOpts); err != nil {
		return err
	}
	// Verify and set standard tls options
	if err := o.setTLSOpts(&serverOpts); err != nil {
		return err
	}
	// Gather options
	o.serverOpts = &serverOpts
	if o.Metrics == nil {
		return nil
	}
	// Initialize exporter options
	exporterOpts := natsmetrics.CollectorOptions{}
	// Verify that monitoring is enabled
	if err := o.verifyMonitoringEnabled(); err != nil {
		return err
	}
	// Verify that the server label is set
	if err := o.setServerLabel(&exporterOpts); err != nil {
		return err
	}
	// Verify that the server URL is set
	if err := o.setServerUrl(&exporterOpts); err != nil {
		return err
	}
	// Enable flags
	if err := o.setCollectorFlags(&exporterOpts); err != nil {
		return err
	}
	o.collectorOpts = &exporterOpts
	return nil
}

// Server returns a NATS server.
func (o *Options) CreateServer() (*server.Server, error) {
	if o.serverOpts == nil {
		return nil, errors.New("server options have not been provisioned yet")
	}
	srv, err := server.NewServer(o.serverOpts)
	if err != nil {
		return nil, err
	}
	natslogger.New(o.logger, o.serverOpts).Attach(srv)
	return srv, nil
}

// Collector returns a NATS metrics collector.
func (o *Options) CreateCollector() (*natsmetrics.Collector, error) {
	if o.collectorOpts == nil {
		return nil, errors.New("collector options have not been provisioned yet")
	}
	return natsmetrics.NewCollector(o.collectorOpts), nil
}

type ServiceImport struct {
	Account string `json:"account"`
	Subject string `json:"subject"`
	To      string `json:"to,omitempty"`
}

type StreamImport struct {
	Account string `json:"account"`
	Subject string `json:"subject"`
	To      string `json:"to,omitempty"`
}

type ServiceExport struct {
	Subject string   `json:"subject"`
	To      []string `json:"to,omitempty"`
}

type Imports struct {
	Services []ServiceImport `json:"services,omitempty"`
	Streams  []StreamImport  `json:"streams,omitempty"`
}

type Exports struct {
	Services []ServiceExport `json:"services,omitempty"`
	Streams  []StreamExport  `json:"streams,omitempty"`
}

type StreamExport struct {
	Subject string   `json:"subject"`
	To      []string `json:"to,omitempty"`
}

type AccountLimits = jwt.AccountLimits

// Account is the configuration for a server account.
// It can be used when defining an authorization configuration.
type Account struct {
	services []natsclient.ServiceProviderModule

	Name                  string                          `json:"name,omitempty"`
	NKey                  string                          `json:"nkey,omitempty"`
	Users                 []*User                         `json:"users,omitempty"`
	JetStream             bool                            `json:"jetstream,omitempty"`
	DefaultPermissions    *server.Permissions             `json:"default_permissions,omitempty"`
	Mappings              []*SubjectMapping               `json:"mappings,omitempty"`
	Limits                *AccountLimits                  `json:"limits,omitempty"`
	Imports               *Imports                        `json:"imports,omitempty"`
	Exports               *Exports                        `json:"exports,omitempty"`
	LeafnodeConnections   []*Remote                       `json:"leafnode_connections,omitempty"`
	AuthorizationPolicies []*natsauth.AuthorizationPolicy `json:"authorization_policies,omitempty"`
	Flows                 []*Flow                         `json:"flows,omitempty"`
	Services              []json.RawMessage               `json:"services,omitempty" caddy:"namespace=nats_server.services inline_key=type"`
	Streams               []*natsclient.Stream            `json:"streams,omitempty"`
	Consumers             []*natsclient.Consumer          `json:"consumers,omitempty"`
	ObjectStores          []*natsclient.ObjectStore       `json:"object_stores,omitempty"`
	KeyValueStores        []*natsclient.KeyValueStore     `json:"key_value_stores,omitempty"`
}

// SubjectMapping is for mapping published subjects for clients.
type SubjectMapping struct {
	Subject string            `json:"subject"`
	MapDest []*server.MapDest `json:"dest"`
}

// AuthCalloutMap is the configuration for the auth callout.
// It can be used when defining an authorization configuration.
type AuthCalloutMap struct {
	Issuer          string                 `json:"issuer,omitempty"`
	SigningKey      string                 `json:"signing_key,omitempty"`
	SigningKeyStore json.RawMessage        `json:"signing_key_store,omitempty" caddy:"namespace=nats_server.signing_key_stores inline_key=type"`
	QueueGroup      string                 `json:"queue_group,omitempty"`
	AuthUsers       []string               `json:"auth_users,omitempty"`
	Account         string                 `json:"account,omitempty"`
	Client          *natsclient.NatsClient `json:"client,omitempty"`
	XKey            string                 `json:"xkey,omitempty"`
}

// User is the configuration for a server user.
// It can be used when defining an authorization configuration.
type User struct {
	User                   string              `json:"user,omitempty"`
	Password               string              `json:"password,omitempty"`
	Permissions            *server.Permissions `json:"permissions,omitempty"`
	AllowedConnectionTypes []string            `json:"allowed_connection_types,omitempty"`
}

// AuthorizationMap block provides authentication configuration
// as well as authorization configuration.
type AuthorizationMap struct {
	Token    string        `json:"token,omitempty"`
	User     string        `json:"user,omitempty"`
	Password string        `json:"password,omitempty"`
	Users    []User        `json:"users,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
}

type ClusterCompression struct {
	Mode string `json:"mode,omitempty"`
	// If `Mode` is set to CompressionS2Auto, RTTThresholds provides the
	// thresholds at which the compression level will go from
	// CompressionS2Uncompressed to CompressionS2Fast, CompressionS2Better
	// or CompressionS2Best. If a given level is not desired, specify 0
	// for this slot. For instance, the slice []{0, 10ms, 20ms} means that
	// for any RTT up to 10ms included the compression level will be
	// CompressionS2Fast, then from ]10ms..20ms], the level will be selected
	// as CompressionS2Better. Anything above 20ms will result in picking
	// the CompressionS2Best compression level.
	RTTThresholds []time.Duration `json:"rtt_thresholds,omitempty"`
}

// Cluster is the configuration for the cluster.
// Cluster is disabled by default. It can be enabled
// by providing a Cluster struct with at least a name:
//
//	Cluster{Name: "cluster-name"}
//
// Default values will be used for missing fields.
// Documentation: https://docs.nats.io/running-a-nats-service/configuration/clustering/cluster_config
type Cluster struct {
	Name           string              `json:"name"`
	Host           string              `json:"host,omitempty"`
	Port           int                 `json:"port,omitempty"`
	Advertise      string              `json:"advertise,omitempty"`
	Routes         []string            `json:"routes,omitempty"`
	TLS            *TLSMap             `json:"tls,omitempty"`
	NoTLS          bool                `json:"no_tls,omitempty"`
	NoAdvertise    bool                `json:"no_advertise,omitempty"`
	Authorization  *AuthorizationMap   `json:"authorization,omitempty"`
	ConnectRetries int                 `json:"connect_retries,omitempty"`
	PoolSize       int                 `json:"pool_size,omitempty"`
	Compression    *ClusterCompression `json:"compression,omitempty"`
}

// Websocket is the configuration for the websocket server
// Websocket server is disabled by default. It can be enabled
// by providing an empty Websocket struct (Websocket{}).
// Default values will be used for missing fields.
type Websocket struct {
	Host           string   `json:"host,omitempty"`
	Port           int      `json:"port,omitempty"`
	Advertise      string   `json:"advertise,omitempty"`
	NoTLS          bool     `json:"no_tls,omitempty"`
	Username       string   `json:"username,omitempty"`
	Password       string   `json:"password,omitempty"`
	NoAuthUser     string   `json:"no_auth_user,omitempty"`
	Compression    bool     `json:"compression,omitempty"`
	SameOrigin     bool     `json:"same_origin,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	JWTCookie      string   `json:"jwt_cookie,omitempty"`
	TLS            *TLSMap  `json:"tls,omitempty"`
}

// MQTT is the configuration for the MQTT server
// MQTT server is disabled by default. It can be enabled
// by providing an empty MQTT struct (MQTT{}).
// Default values will be used for missing fields.
type MQTT struct {
	Host           string  `json:"host,omitempty"`
	Port           int     `json:"port,omitempty"`
	JSDomain       string  `json:"jetstream_domain,omitempty"`
	StreamReplicas int     `json:"stream_replicas,omitempty"`
	Username       string  `json:"username,omitempty"`
	Password       string  `json:"password,omitempty"`
	NoAuthUser     string  `json:"no_auth_user,omitempty"`
	AuthTimeout    float64 `json:"auth_timeout,omitempty"`
	NoTLS          bool    `json:"no_tls,omitempty"`
	TLS            *TLSMap `json:"tls,omitempty"`
}

// JetStream is the configuration for the JetStream engine
// JetStream is disabled by default. It can be enabled
// by providing an empty JetStream struct (JetStream{}).
// Default values will be used for missing fields.
type JetStream struct {
	Domain    string `json:"domain,omitempty"`
	UniqueTag string `json:"unique_tag,omitempty"`
	StoreDir  string `json:"store_dir,omitempty"`
	MaxMemory int64  `json:"max_memory,omitempty"`
	MaxFile   int64  `json:"max_file,omitempty"`
}

// RemoteWebsocketClient is the configuration for the websocket client connection
// used to connect to a remote leafnode.
type RemoteWebsocketClient struct {
	Compression bool `json:"compression,omitempty"`
	NoMasking   bool `json:"no_masking,omitempty"`
}

// Remote is the configuration for remote leafnode connections.
// It is used when defining a leafnode configuration.
// Default values will be used for missing fields.
// Either a single URL can be provided, or a list of URLs.
type Remote struct {
	Urls        []string               `json:"urls,omitempty"`
	Hub         bool                   `json:"hub,omitempty"`
	DenyImports []string               `json:"deny_imports,omitempty"`
	DenyExports []string               `json:"deny_exports,omitempty"`
	NoRandomize bool                   `json:"no_randomize,omitempty"`
	Account     string                 `json:"account,omitempty"`
	Credentials string                 `json:"credentials,omitempty"`
	Websocket   *RemoteWebsocketClient `json:"websocket,omitempty"`
}

// Leafnode is the configuration for the leafnode server
// Leafnode server is disabled by default. It can be enabled
// by providing an empty Leafnode struct (Leafnode{}).
// Default values will be used for missing fields.
type Leafnode struct {
	Host      string    `json:"host,omitempty"`
	Port      int       `json:"port,omitempty"`
	Advertise string    `json:"advertise,omitempty"`
	NoTLS     bool      `json:"no_tls,omitempty"`
	TLS       *TLSMap   `json:"tls,omitempty"`
	Remotes   []*Remote `json:"remotes,omitempty"`
}

// FullAccountResolver is the configuration for the full NATS account resolver.
// Account resolver is disabled by default. It can be enabled
// by providing an empty FullAccountResolver struct (FullAccountResolver{}).
// Default values will be used for missing fields.
type FullAccountResolver struct {
	Path         string        `json:"path,omitempty"`
	Limit        int64         `json:"limit,omitempty"`
	SyncInterval time.Duration `json:"interval,omitempty"`
	AllowDelete  bool          `json:"allow_delete,omitempty"`
	HardDelete   bool          `json:"hard_delete,omitempty"`
	Preload      []string      `json:"preload,omitempty"`
}

type CacheAccountResolver struct {
	Path    string        `json:"path,omitempty"`
	Limit   int           `json:"limit,omitempty"`
	TTL     time.Duration `json:"ttl,omitempty"`
	Preload []string      `json:"preload,omitempty"`
}

type MemoryAccountResolver struct {
	Limit   int      `json:"limit,omitempty"`
	Preload []string `json:"preload,omitempty"`
}

// Metrics is the configuration for the prometheus metrics
// collector.
// Metrics collector is disabled by default. It can be enabled
// by providing an empty Metrics struct (Metrics{}).
// Only /varz metrics are enabled by defaul, other metrics must be enabled
// explicitly.
type Metrics struct {
	ServerLabel   string `json:"server_label,omitempty"`
	ServerUrl     string `json:"server_url,omitempty"`
	Healthz       bool   `json:"healthz,omitempty"`
	Connz         bool   `json:"connz,omitempty"`
	ConnzDetailed bool   `json:"connz_detailed,omitempty"`
	Subz          bool   `json:"subz,omitempty"`
	Routez        bool   `json:"routez,omitempty"`
	Gatewayz      bool   `json:"gatewayz,omitempty"`
	Leafz         bool   `json:"leafz,omitempty"`
	JszFilter     string `json:"jsz_filter,omitempty"`
}

// TLSMap is a configuration block for TLSMap servers.
type TLSMap struct {
	AutomationPolicy *caddytls.AutomationPolicy `json:"automation_policy,omitempty"`
	Subjects         []string                   `json:"subjects,omitempty"`
	AllowNonTLS      bool                       `json:"allow_non_tls,omitempty"`
	CertFile         string                     `json:"cert_file,omitempty"`
	CertStore        string                     `json:"cert_store,omitempty"`
	CertMatch        string                     `json:"cert_match,omitempty"`
	CertMatchBy      string                     `json:"cert_match_by,omitempty"`
	KeyFile          string                     `json:"key_file,omitempty"`
	CaFile           string                     `json:"ca_file,omitempty"`
	Verify           bool                       `json:"verify,omitempty"`
	Insecure         bool                       `json:"insecure,omitempty"`
	Map              bool                       `json:"map,omitempty"`
	CheckKnownURLs   bool                       `json:"tls_check_known_urls,omitempty"`
	Timeout          time.Duration              `json:"timeout,omitempty"`
	RateLimit        int64                      `json:"rate_limit,omitempty"`
	Ciphers          []string                   `json:"ciphers,omitempty"`
	CurvePreferences []string                   `json:"curve_preferences,omitempty"`
	PinnedCerts      []string                   `json:"pinned_certs,omitempty"`
}

func (t *TLSMap) getConfig(o *Options) (*tls.Config, error) {
	if t.Subjects != nil || t.AutomationPolicy != nil {
		var subjects = t.Subjects
		if o.tlsApp == nil {
			unm, err := o.ctx.App("tls")
			if err != nil {
				return nil, err
			}
			tlsapp, ok := unm.(*caddytls.TLS)
			if !ok {
				return nil, fmt.Errorf("tls app is not a caddytls.TLS")
			}
			o.tlsApp = tlsapp
		}
		if subjects != nil && t.AutomationPolicy != nil {
			if t.AutomationPolicy.SubjectsRaw != nil {
				return nil, fmt.Errorf("cannot specify both 'subjects' and 'automation_policy.subjects'")
			}
			t.AutomationPolicy.SubjectsRaw = subjects
		}
		if t.AutomationPolicy == nil {
			t.AutomationPolicy = &caddytls.AutomationPolicy{
				SubjectsRaw:         subjects,
				IssuersRaw:          o.AutomationPolicyTemplate.IssuersRaw,
				ManagersRaw:         o.AutomationPolicyTemplate.ManagersRaw,
				MustStaple:          o.AutomationPolicyTemplate.MustStaple,
				RenewalWindowRatio:  o.AutomationPolicyTemplate.RenewalWindowRatio,
				KeyType:             o.AutomationPolicyTemplate.KeyType,
				StorageRaw:          o.AutomationPolicyTemplate.StorageRaw,
				OnDemand:            o.AutomationPolicyTemplate.OnDemand,
				DisableOCSPStapling: o.AutomationPolicyTemplate.DisableOCSPStapling,
				OCSPOverrides:       o.AutomationPolicyTemplate.OCSPOverrides,
			}
		}
		if err := o.tlsApp.AddAutomationPolicy(t.AutomationPolicy); err != nil {
			return nil, err
		}
		subjects = t.AutomationPolicy.Subjects()
		policies := caddytls.ConnectionPolicies{
			&caddytls.ConnectionPolicy{
				MatchersRaw: map[string]json.RawMessage{
					"sni": caddyconfig.JSON(subjects, nil),
				},
			},
		}
		if err := policies.Provision(o.ctx); err != nil {
			return nil, err
		}
		config := policies.TLSConfig(o.ctx)
		// Override GetCertificate
		cfg := config.Clone()
		// Reset certificates
		cfg.Certificates = nil
		// TODO: Add a comment and reference NATS code which make this necessary
		cfg.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			_cfg, err := config.GetConfigForClient(hello)
			if err != nil {
				return nil, err
			}
			return _cfg.GetCertificate(hello)
		}
		cfg.GetConfigForClient = func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			_cfg, err := config.GetConfigForClient(hello)
			if err != nil {
				return nil, err
			}
			return _cfg, nil
		}
		o.automateSubjects = append(o.automateSubjects, subjects...)
		return cfg, nil
	}
	// Parse ciphers
	var ciphers = []uint16{}
	for _, name := range t.Ciphers {
		cipher, err := natstls.ParseCipherFromName(name)
		if err != nil {
			return nil, err
		}
		ciphers = append(ciphers, cipher)
	}
	// Use default cipher suites if none specified
	if len(ciphers) == 0 {
		ciphers = natstls.DefaultCipherSuites()
	}
	// Parse curves preferences
	var curves = []tls.CurveID{}
	for _, curve := range t.CurvePreferences {
		if curve == "" {
			continue
		}
		curveID, ok := natstls.CurvePreferenceMap[curve]
		if !ok {
			return nil, fmt.Errorf("invalid tls curve preference: %s", curve)
		}
		curves = append(curves, curveID)
	}
	// Use default curve preferences if none specified
	if len(curves) == 0 {
		curves = natstls.DefaultCurvePreferences()
	}
	// Create the tls.Config from our options before including the certs.
	// It will determine the cipher suites that we prefer.
	config := tls.Config{
		MinVersion:       tls.VersionTLS12,
		CipherSuites:     ciphers,
		CurvePreferences: curves,
	}
	switch {
	case t.CertFile != "" && t.CertStore != "":
		return nil, certstore.ErrConflictCertFileAndStore
	case t.CertFile != "" && t.KeyFile == "":
		return nil, fmt.Errorf("missing 'key_file' in TLS configuration")
	case t.CertFile == "" && t.KeyFile != "":
		return nil, fmt.Errorf("missing 'cert_file' in TLS configuration")
	case t.CertFile != "" && t.KeyFile != "":
		// Now load in cert and private key
		cert, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	case t.CertStore != "":
		store, err := certstore.ParseCertStore(t.CertStore)
		if err != nil {
			return nil, fmt.Errorf("invalid tls.cert_store option: %s", err.Error())
		}
		matchBy, err := certstore.ParseCertMatchBy(t.CertMatchBy)
		if err != nil {
			return nil, fmt.Errorf("invalid tls.cert_match_by option: %s", err.Error())
		}
		err = certstore.TLSConfig(store, matchBy, t.CertMatch, &config)
		if err != nil {
			return nil, fmt.Errorf("error generating tls config using cert_store: %s", err.Error())
		}
	}
	// Require client certificates as needed
	if t.Verify {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}
	// Add in CAs if applicable.
	if t.CaFile != "" {
		rootPEM, err := os.ReadFile(t.CaFile)
		if err != nil || rootPEM == nil {
			return nil, fmt.Errorf("error reading tls root ca certificate: %s", err.Error())
		}
		pool := x509.NewCertPool()
		ok := pool.AppendCertsFromPEM(rootPEM)
		if !ok {
			return nil, fmt.Errorf("error parsing tls root ca certificate")
		}
		config.ClientCAs = pool
	}
	return &config, nil
}

func (o *Options) setGlobalOpts(opts *server.Options) error {
	opts.ServerName = o.ServerName
	tags := []string{}
	for key, value := range o.ServerTags {
		tag := strings.Join([]string{key, value}, ":")
		tags = append(tags, tag)
	}
	opts.Tags = tags
	opts.Host = o.Host
	opts.Port = o.Port
	opts.ClientAdvertise = o.Advertise
	opts.Debug = o.Debug
	opts.Trace = o.Trace
	opts.TraceVerbose = o.TraceVerbose
	opts.NoLog = o.NoLog
	opts.NoSublistCache = o.NoSublistCache
	opts.MaxConn = o.MaxConn
	opts.MaxSubs = o.MaxSubs
	opts.MaxPayload = o.MaxPayload
	opts.MaxPending = o.MaxPending
	opts.MaxClosedClients = o.MaxClosedClients
	opts.MaxControlLine = o.MaxControlLine
	opts.MaxPingsOut = o.MaxPingsOut
	opts.MaxSubTokens = o.MaxSubsTokens
	opts.PingInterval = o.PingInterval
	opts.MaxTracedMsgLen = o.MaxTracedMsgLen
	opts.WriteDeadline = o.WriteDeadline
	return nil
}

func (o *Options) setTLSOpts(opts *server.Options) error {
	if o.TLS == nil {
		return nil
	}
	t := o.TLS
	// Set tls global options (not sure this is useful)
	opts.TLSVerify = t.Verify
	opts.TLSMap = t.Map
	if len(t.PinnedCerts) > 0 {
		certs := map[string]struct{}{}
		for _, cert := range t.PinnedCerts {
			certs[cert] = struct{}{}
		}
		opts.TLSPinnedCerts = certs
	}
	opts.AllowNonTLS = t.AllowNonTLS
	opts.TLSRateLimit = t.RateLimit
	opts.TLSTimeout = t.Timeout.Seconds()
	config, err := t.getConfig(o)
	if err != nil {
		return err
	}
	opts.TLSConfig = config
	return nil
}

func (o *Options) setResolverOpts(opts *server.Options) error {
	if o.FullResolver == nil && o.CacheResolver == nil && o.MemoryResolver == nil {
		if o.Operators != nil {
			return errors.New("operators are set but resolver is not configured")
		}
		return nil
	}
	if o.FullResolver != nil {
		if o.MemoryResolver != nil || o.CacheResolver != nil {
			return errors.New("full_resolver and memory_resolver/cache_resolver cannot be set at the same time")
		}
		var deleteType = server.NoDelete
		if o.FullResolver.AllowDelete && o.FullResolver.HardDelete {
			deleteType = server.HardDelete
		} else if o.FullResolver.AllowDelete {
			deleteType = server.RenameDeleted
		}

		resolver, err := server.NewDirAccResolver(o.FullResolver.Path, o.FullResolver.Limit, o.FullResolver.SyncInterval, deleteType)
		if err != nil {
			return fmt.Errorf("invalid full resolver: %s", err.Error())
		}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.FullResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = resolver
	}
	if o.MemoryResolver != nil {
		if o.CacheResolver != nil || o.FullResolver != nil {
			return errors.New("memory_resolver and cache_resolver/full_resolver cannot be set at the same time")
		}
		resolver := server.MemAccResolver{}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.MemoryResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = &resolver
	}
	if o.CacheResolver != nil {
		if o.MemoryResolver != nil || o.FullResolver != nil {
			return errors.New("cache_resolver and memory_resolver/full_resolver cannot be set at the same time")
		}
		resolver, err := server.NewCacheDirAccResolver(
			o.CacheResolver.Path,
			int64(o.CacheResolver.Limit),
			o.CacheResolver.TTL,
		)
		if err != nil {
			return fmt.Errorf("invalid cache resolver: %s", err.Error())
		}
		if o.SystemAccount != "" && o.systemAccount != nil {
			if err := resolver.Store(o.systemAccount.Subject, o.SystemAccount); err != nil {
				return fmt.Errorf("invalid system account: %s", err.Error())
			}
		}
		for _, entry := range o.CacheResolver.Preload {
			claims, err := jwt.DecodeAccountClaims(entry)
			if err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
			if err := resolver.Store(claims.Subject, entry); err != nil {
				return fmt.Errorf("invalid memory resolver preload entry: %s", err.Error())
			}
		}
		opts.AccountResolver = resolver
	}
	return nil
}

func (o *Options) setSystemAccountOpt(opts *server.Options) error {
	if o.Operators != nil {
		// Parse system account jwt
		claims, err := jwt.DecodeAccountClaims(o.SystemAccount)
		if err != nil {
			return fmt.Errorf("invalid system account: %s", err.Error())
		}
		opts.SystemAccount = claims.Subject
		o.systemAccount = claims
		return nil
	}
	// Don't attempt to parse, system account may be a simple name, maybe it's empty
	opts.SystemAccount = o.SystemAccount
	// Check is system account must be created
	if o.SystemAccount == "" && o.systemAccount == nil && o.Accounts != nil {
		// We have accounts, but we don't have a system account.
		// Let's create one named "SYS"
		o.SystemAccount = "SYS"
		// If this account already exists, raise an error, because we don't know
		// if administrator is aware that this will be the system account or not
		for _, account := range o.Accounts {
			if account.Name == o.SystemAccount {
				return errors.New("system account must be explicitely specified when an account named SYS is used")
			}
		}
		o.addAccount(opts, &Account{Name: o.SystemAccount})
	}
	return nil
}

func (o *Options) setOperatorModeOpts(opts *server.Options) error {
	if o.Operators == nil {
		return nil
	}
	operators := []*jwt.OperatorClaims{}
	for _, token := range o.Operators {
		claims, err := jwt.DecodeOperatorClaims(token)
		if err != nil {
			return fmt.Errorf("invalid operator token: %s", err.Error())
		}
		// Set default system account
		if opts.SystemAccount == "" && claims.SystemAccount != "" {
			opts.SystemAccount = claims.SystemAccount
		}
		operators = append(operators, claims)
	}
	opts.TrustedOperators = operators
	return nil
}

func (o *Options) setTokenAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.Token == "" {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.token and operators cannot be set at the same time")
	}
	if o.Authorization.Users != nil {
		return errors.New("authorization.token and authorization.users cannot be set at the same time")
	}
	if o.Authorization.User != "" {
		return errors.New("authorization.token and authorization.user cannot be set at the same time")
	}
	opts.Authorization = o.Authorization.Token
	return nil
}

func (o *Options) setUserPasswordAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.User == "" {
		if o.Authorization.Password != "" {
			return errors.New("authorization.password cannot be set without authorization.user")
		}
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.user and operators cannot be set at the same time")
	}
	if o.Authorization.Users != nil {
		return errors.New("authorization.user and authorization.users cannot be set at the same time")
	}
	opts.Username = o.Authorization.User
	opts.Password = o.Authorization.Password
	return nil
}

func (o *Options) addUser(opts *server.Options, user *User) error {
	if user.User == "" {
		return errors.New("cannot add user without a name")
	}
	if user.Password == "" {
		return errors.New("cannot add user without a password")
	}
	allowedConnTypes, err := validateConnectionTypes(user.AllowedConnectionTypes)
	if err != nil {
		return err
	}
	opts.Users = append(opts.Users, &server.User{
		Username:               user.User,
		Password:               user.Password,
		Permissions:            user.Permissions,
		AllowedConnectionTypes: allowedConnTypes,
	})
	return nil
}

func (o *Options) setUsersAuth(opts *server.Options) error {
	if o.Authorization == nil {
		return nil
	}
	if o.Authorization.Users == nil {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.users and operators cannot be set at the same time")
	}
	if o.Authorization.Users == nil {
		return nil
	}
	if len(o.Authorization.Users) == 0 {
		return errors.New("authorization.users must either be omitted or set with at least one user")
	}
	opts.Users = []*server.User{}
	for _, user := range o.Authorization.Users {
		if err := o.addUser(opts, &user); err != nil {
			return fmt.Errorf("invalid user: %s", err.Error())
		}
	}
	return nil
}

func (o *Options) addAccountUser(opts *server.Options, account *server.Account, user *User) error {
	if user.User == "" {
		return errors.New("cannot add an account user without a name")
	}
	if user.Password == "" {
		return errors.New("cannot add an account user without a password")
	}
	allowedConnTypes, err := validateConnectionTypes(user.AllowedConnectionTypes)
	if err != nil {
		return err
	}
	accUser := server.User{
		Username:               user.User,
		Password:               user.Password,
		Permissions:            user.Permissions,
		AllowedConnectionTypes: allowedConnTypes,
		Account:                account,
	}
	opts.Users = append(opts.Users, &accUser)
	return nil
}

func (o *Options) addAccount(opts *server.Options, account *Account) error {
	if account.Name == "" {
		return errors.New("authorization.accounts.name cannot be empty")
	}
	acc := server.NewAccount(account.Name)
	// Add mappings
	for _, mapping := range account.Mappings {
		if err := acc.AddWeightedMappings(mapping.Subject, mapping.MapDest...); err != nil {
			return fmt.Errorf("invalid account subject mapping: %s", err.Error())
		}
	}
	// Add users
	for _, user := range account.Users {
		if err := o.addAccountUser(opts, acc, user); err != nil {
			return fmt.Errorf("invalid user: %s", err.Error())
		}
	}
	// Add authorization policies
	for _, pol := range account.AuthorizationPolicies {
		if err := pol.Provision(o.ctx, account.Name); err != nil {
			return fmt.Errorf("failed to provision authorization policy: %s", err.Error())
		}
	}
	// Add services
	unm, err := o.ctx.LoadModule(account, "Services")
	if err != nil {
		return fmt.Errorf("invalid services: %s", err.Error())
	}
	for _, raw := range unm.([]interface{}) {
		svc, ok := raw.(natsclient.ServiceProviderModule)
		if !ok {
			return errors.New("invalid service")
		}
		if err := svc.Provision(o.ctx); err != nil {
			return fmt.Errorf("failed to provision service: %s", err.Error())
		}
		account.services = append(account.services, svc)
	}
	opts.Accounts = append(opts.Accounts, acc)
	return nil
}

func (o *Options) setAuthCallout(opts *server.Options) error {
	if o.AuthCallout == nil {
		return nil
	}
	acc := o.AuthCallout.Account
	if acc == "" {
		switch {
		case len(opts.Accounts) == 0:
			acc = "$G"
		default:
			acc = "AUTH"
		}
	}
	if o.AuthCallout.SigningKey != "" && o.AuthCallout.Issuer != "" {
		return errors.New("auth_callout.signing_key and auth_callout.issuer cannot be set at the same time")
	}
	if o.AuthCallout.SigningKeyStore != nil && o.AuthCallout.Issuer == "" {
		return errors.New("auth_callout.signing_key_store and auth_callout.issuer must be set at the same time")
	}
	var store KeyStore
	var privateKey nkeys.KeyPair
	var publicKey = o.AuthCallout.Issuer
	if o.AuthCallout.SigningKeyStore != nil {
		unm, err := o.ctx.LoadModule(o.AuthCallout, "SigningKeyStore")
		if err != nil {
			return fmt.Errorf("invalid auth_callout.signing_key_store: %s", err.Error())
		}
		keystore, ok := unm.(KeyStore)
		if !ok {
			return errors.New("invalid auth_callout.signing_key_store")
		}
		if err := keystore.Provision(o.ctx); err != nil {
			return fmt.Errorf("failed to provision auth_callout.signing_key_store: %s", err.Error())
		}
		store = keystore
	}
	if o.AuthCallout.SigningKey == "" && o.AuthCallout.Issuer == "" {
		if __authsigningkey__ != nil {
			privateKey = __authsigningkey__
			publicKey = __authpublickey__
		} else {
			keypair, err := nkeys.CreateAccount()
			if err != nil {
				return fmt.Errorf("failed to generate signing key: %s", err.Error())
			}
			pk, err := keypair.PublicKey()
			if err != nil {
				return fmt.Errorf("failed to generate signing key: %s", err.Error())
			}
			__authsigningkey__ = keypair
			__authpublickey__ = pk
			privateKey = keypair
			publicKey = pk
		}
	}
	if o.AuthCallout.SigningKey != "" {
		keypair, err := nkeys.FromSeed([]byte(o.AuthCallout.SigningKey))
		if err != nil {
			return fmt.Errorf("invalid auth_callout.signing_key: %s", err.Error())
		}
		pk, err := keypair.PublicKey()
		if err != nil {
			return fmt.Errorf("invalid auth_callout.signing_key: %s", err.Error())
		}
		privateKey = keypair
		publicKey = pk
	}
	opts.AuthCallout = &server.AuthCallout{
		Account:   acc,
		Issuer:    publicKey,
		AuthUsers: o.AuthCallout.AuthUsers,
		XKey:      o.AuthCallout.XKey,
	}
	if privateKey != nil || store != nil {
		o.authService = &authService{
			ctx:        o.ctx,
			logger:     o.logger.Named("auth_service"),
			queueGroup: o.AuthCallout.QueueGroup,
			account:    acc,
			issuer:     publicKey,
			policies:   natsauth.AuthorizationPolicies{},
		}
		if privateKey != nil {
			o.authService.signingKey = privateKey
		}
		if store != nil {
			o.authService.keystore = store
		}
	}
	return nil
}

func (o *Options) setAccountsAuth(opts *server.Options) error {
	if o.Accounts == nil {
		return nil
	}
	if o.Operators != nil {
		return errors.New("authorization.accounts and operators cannot be set at the same time")
	}
	if o.Authorization != nil {
		return errors.New("authorization.accounts and authorization.users cannot be set at the same time")
	}
	if len(o.Accounts) == 0 {
		return errors.New("authorization.accounts must either be omitted or set with at least one account")
	}
	opts.Accounts = []*server.Account{}
	for _, account := range o.Accounts {
		if err := o.addAccount(opts, account); err != nil {
			return fmt.Errorf("invalid account: %s", err.Error())
		}
	}
	return nil
}

func (o *Options) setNoAuthUser(opts *server.Options) error {
	if o.NoAuthUser != "" {
		if o.Operators != nil {
			return errors.New("no_auth_user and operators cannot be set at the same time")
		}
		opts.NoAuthUser = o.NoAuthUser
	}
	return nil
}

func (o *Options) setMonitoringOpts(opts *server.Options) error {
	// Verify that one of http_port or https_port may be defined but not both
	if o.HTTPPort != 0 && o.HTTPSPort != 0 {
		return errors.New("metrics.http_port and metrics.https_port cannot be set at the same time")
	}
	opts.HTTPPort = o.HTTPPort
	opts.HTTPSPort = o.HTTPSPort
	opts.HTTPHost = o.HTTPHost
	opts.HTTPBasePath = o.HTTPBasePath
	return nil
}

func (o *Options) setJetStreamOpts(opts *server.Options) error {
	if o.JetStream == nil {
		return nil
	}
	opts.JetStream = true
	opts.JetStreamMaxMemory = o.JetStream.MaxMemory
	opts.JetStreamMaxStore = o.JetStream.MaxFile
	opts.StoreDir = o.JetStream.StoreDir
	opts.JetStreamDomain = o.JetStream.Domain
	opts.JetStreamUniqueTag = o.JetStream.UniqueTag
	return nil
}

func (o *Options) setMqttOpts(opts *server.Options) error {
	if o.Mqtt == nil {
		return nil
	}
	if o.JetStream == nil {
		return errors.New("mqtt cannot be enabled without jetstream")
	}
	if o.ServerName == "" {
		return errors.New("mqtt cannot be enabled without server name")
	}
	// Set default port if none specified
	var port = o.Mqtt.Port
	if port == 0 {
		if o.Mqtt.TLS != nil {
			port = 8883
		} else {
			port = 1883
		}
	}
	opts.MQTT.Host = o.Mqtt.Host
	opts.MQTT.Port = port
	opts.MQTT.Username = o.Mqtt.Username
	opts.MQTT.Password = o.Mqtt.Password
	opts.MQTT.AuthTimeout = o.Mqtt.AuthTimeout
	opts.MQTT.StreamReplicas = o.Mqtt.StreamReplicas
	opts.MQTT.NoAuthUser = o.Mqtt.NoAuthUser
	return nil
}

func (o *Options) setWebsocketOpts(opts *server.Options) error {
	if o.Websocket == nil {
		return nil
	}
	var port = o.Websocket.Port
	if port == 0 {
		if o.Websocket.TLS != nil {
			port = 10443
		} else {
			port = 10080
		}
	}
	opts.Websocket.Host = o.Websocket.Host
	opts.Websocket.Port = port
	opts.Websocket.Advertise = o.Websocket.Advertise
	opts.Websocket.NoTLS = o.Websocket.NoTLS
	opts.Websocket.Username = o.Websocket.Username
	opts.Websocket.Password = o.Websocket.Password
	opts.Websocket.NoAuthUser = o.Websocket.NoAuthUser
	opts.Websocket.Compression = o.Websocket.Compression
	opts.Websocket.SameOrigin = o.Websocket.SameOrigin
	opts.Websocket.AllowedOrigins = o.Websocket.AllowedOrigins
	opts.Websocket.JWTCookie = o.Websocket.JWTCookie
	if o.Websocket.TLS != nil {
		t := o.Websocket.TLS
		// Verify and set websocket tls options
		// Set tls global options (not sure this is useful)
		opts.Websocket.TLSMap = t.Map
		if len(t.PinnedCerts) > 0 {
			certs := map[string]struct{}{}
			for _, cert := range t.PinnedCerts {
				certs[cert] = struct{}{}
			}
			opts.Websocket.TLSPinnedCerts = certs
		}
		// If TLS is managed, use the provided tls.Config
		config, err := t.getConfig(o)
		if err != nil {
			return err
		}
		opts.Websocket.TLSConfig = config
	}
	return nil
}

func (o *Options) setLeafnodeOpts(opts *server.Options) error {
	// Override leafnode connections using accounts
	for _, account := range o.Accounts {
		if account.LeafnodeConnections == nil {
			continue
		}
		for _, remote := range account.LeafnodeConnections {
			urls := []*url.URL{}
			for _, r := range remote.Urls {
				parsed, err := url.Parse(r)
				if err != nil {
					return fmt.Errorf("invalid leafnode connection url: %s", err.Error())
				}
				urls = append(urls, parsed)
			}
			serverLeaf := &server.RemoteLeafOpts{
				URLs:         urls,
				NoRandomize:  remote.NoRandomize,
				LocalAccount: remote.Account,
				Hub:          remote.Hub,
				Credentials:  remote.Credentials,
				DenyImports:  remote.DenyImports,
				DenyExports:  remote.DenyExports,
			}
			if remote.Websocket != nil {
				serverLeaf.Websocket = struct {
					Compression bool `json:"-"`
					NoMasking   bool `json:"-"`
				}(*remote.Websocket)
			}
			if opts.LeafNode.Remotes == nil {
				opts.LeafNode.Remotes = []*server.RemoteLeafOpts{}
			}
			opts.LeafNode.Remotes = append(opts.LeafNode.Remotes, serverLeaf)
		}
	}
	if o.Leafnode == nil {
		return nil
	}
	port := o.Leafnode.Port
	// Set default listening port when no remotes are defined
	if port == 0 && len(o.Leafnode.Remotes) == 0 {
		port = 7422
	}
	opts.LeafNode.Host = o.Leafnode.Host
	opts.LeafNode.Port = port
	opts.LeafNode.Advertise = o.Leafnode.Advertise
	if opts.LeafNode.Remotes == nil {
		opts.LeafNode.Remotes = []*server.RemoteLeafOpts{}
	}
	opts.LeafNode.Remotes = []*server.RemoteLeafOpts{}
	for _, remote := range o.Leafnode.Remotes {
		urls := []*url.URL{}
		for _, r := range remote.Urls {
			parsed, err := url.Parse(r)
			if err != nil {
				return fmt.Errorf("invalid remote leafnode url: %s", err.Error())
			}
			urls = append(urls, parsed)
		}
		serverLeaf := &server.RemoteLeafOpts{
			URLs:         urls,
			NoRandomize:  remote.NoRandomize,
			LocalAccount: remote.Account,
			Hub:          remote.Hub,
			Credentials:  remote.Credentials,
			DenyImports:  remote.DenyImports,
			DenyExports:  remote.DenyExports,
		}
		if remote.Websocket != nil {
			serverLeaf.Websocket = struct {
				Compression bool `json:"-"`
				NoMasking   bool `json:"-"`
			}(*remote.Websocket)
		}
		opts.LeafNode.Remotes = append(opts.LeafNode.Remotes, serverLeaf)
	}
	if o.Leafnode.TLS != nil {
		t := o.Leafnode.TLS
		// Set tls global options (not sure this is useful)
		opts.LeafNode.TLSMap = t.Map
		if len(t.PinnedCerts) > 0 {
			certs := map[string]struct{}{}
			for _, cert := range t.PinnedCerts {
				certs[cert] = struct{}{}
			}
			opts.LeafNode.TLSPinnedCerts = certs
		}
		opts.LeafNode.TLSTimeout = t.Timeout.Seconds()
		config, err := t.getConfig(o)
		if err != nil {
			return err
		}
		opts.LeafNode.TLSConfig = config
	}
	return nil
}

func (o *Options) setClusterOpts(opts *server.Options) error {
	if o.Cluster == nil {
		return nil
	}
	if o.Cluster.Name == "" {
		return errors.New("cluster.name cannot be empty")
	}
	port := o.Cluster.Port
	if port == 0 {
		port = 6222
	}
	opts.Cluster.Name = o.Cluster.Name
	opts.Cluster.Host = o.Cluster.Host
	opts.Cluster.Port = port
	opts.Cluster.Advertise = o.Cluster.Advertise
	opts.Cluster.NoAdvertise = o.Cluster.NoAdvertise
	opts.Cluster.ConnectRetries = o.Cluster.ConnectRetries
	opts.Cluster.PoolSize = o.Cluster.PoolSize
	if o.Cluster.Compression != nil {
		opts.Cluster.Compression = server.CompressionOpts{Mode: o.Cluster.Compression.Mode, RTTThresholds: o.Cluster.Compression.RTTThresholds}
	}
	if len(o.Cluster.Routes) == 0 {
		opts.Routes = make([]*url.URL, 1)
		routeUrl, err := url.Parse(fmt.Sprintf("nats-route://localhost:%d", port))
		if err != nil {
			return fmt.Errorf("invalid cluster route url: %s", err.Error())
		}
		opts.Routes[0] = routeUrl
	} else {
		opts.Routes = make([]*url.URL, len(o.Cluster.Routes))
		for i, route := range o.Cluster.Routes {
			routeUrl, err := url.Parse(route)
			if err != nil {
				return fmt.Errorf("invalid cluster route url: %s", err.Error())
			}
			opts.Routes[i] = routeUrl
		}
	}
	return nil
}

func (o *Options) verifyMonitoringEnabled() error {
	if o.HTTPPort == 0 && o.HTTPSPort == 0 {
		return errors.New("metrics.http_port or metrics.https_port must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setServerLabel(opts *natsmetrics.CollectorOptions) error {
	// Verify that the server label is set
	if o.Metrics.ServerLabel != "" {
		opts.NATSServerLabel = o.Metrics.ServerLabel
	} else if o.ServerName != "" {
		opts.NATSServerLabel = o.ServerName
	} else {
		return errors.New("metrics.server_label or server_name must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setServerUrl(opts *natsmetrics.CollectorOptions) error {
	// Verify that the server URL is set
	if o.Metrics.ServerUrl != "" {
		opts.NATSServerURL = o.Metrics.ServerUrl
	} else if o.HTTPPort != 0 {
		opts.NATSServerURL = fmt.Sprintf("http://localhost:%d", o.HTTPPort)
	} else if o.HTTPSPort != 0 {
		return errors.New("metrics.server_url must be enabled when metrics.https_port is set")
	} else {
		return errors.New("either http_port or metrics.server_url and https_port must be set when metrics are enabled")
	}
	return nil
}

func (o *Options) setCollectorFlags(opts *natsmetrics.CollectorOptions) error {
	// GetVarz is always enabled
	opts.GetVarz = true
	opts.GetHealthz = o.Metrics.Healthz
	opts.GetConnz = o.Metrics.Connz
	opts.GetConnzDetailed = o.Metrics.ConnzDetailed
	opts.GetSubz = o.Metrics.Subz
	opts.GetRoutez = o.Metrics.Routez
	opts.GetGatewayz = o.Metrics.Gatewayz
	opts.GetLeafz = o.Metrics.Leafz
	opts.GetJszFilter = o.Metrics.JszFilter
	return nil
}

func validateConnectionTypes(allowedConnectionTypes []string) (map[string]struct{}, error) {
	allowed := map[string]struct{}{}
	for _, connType := range allowedConnectionTypes {
		typ := strings.ToUpper(connType)
		if typ != jwt.ConnectionTypeStandard &&
			typ != jwt.ConnectionTypeWebsocket &&
			typ != jwt.ConnectionTypeMqtt &&
			typ != jwt.ConnectionTypeMqttWS &&
			typ != jwt.ConnectionTypeLeafnode &&
			typ != jwt.ConnectionTypeLeafnodeWS {
			return nil, fmt.Errorf("invalid connection type: %q", connType)
		}
		allowed[connType] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	return allowed, nil
}
