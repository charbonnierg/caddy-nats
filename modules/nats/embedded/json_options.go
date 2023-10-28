// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package options provides a JSON struct which can be used to generate a valid
// embedded NATS server configuration.
package embedded

import (
	"crypto/tls"
	"encoding/json"
	"os"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
)

func NewOptions() *Options {
	return &Options{}
}

func NewFromJSON(data []byte) (*Options, error) {
	opts := NewOptions()
	if err := json.Unmarshal(data, opts); err != nil {
		return nil, err
	}
	return opts, nil
}

func NewFromJSONFile(path string) (*Options, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewFromJSON(content)
}

// Options is the configuration for the NATS server
// It can be used to generate a server.Options struct.
// Default values will be used for missing fields.
type Options struct {
	systemAccount    *jwt.AccountClaims
	ServerName       string                 `json:"name,omitempty"`
	ServerTags       map[string]string      `json:"tags,omitempty"`
	Host             string                 `json:"host,omitempty"`
	Port             int                    `json:"port,omitempty"`
	Advertise        string                 `json:"advertise,omitempty"`
	Debug            bool                   `json:"debug,omitempty"`
	Trace            bool                   `json:"trace,omitempty"`
	TraceVerbose     bool                   `json:"trace_verbose,omitempty"`
	HTTPHost         string                 `json:"http_host,omitempty"`
	HTTPPort         int                    `json:"http_port,omitempty"`
	HTTPSPort        int                    `json:"https_port,omitempty"`
	HTTPBasePath     string                 `json:"http_base_path,omitempty"`
	NoLog            bool                   `json:"disable_logging,omitempty"`
	NoTLS            bool                   `json:"no_tls,omitempty"`
	TLS              *TLSMap                `json:"tls,omitempty"`
	NoSublistCache   bool                   `json:"disable_sublist_cache,omitempty"`
	MaxConn          int                    `json:"max_connections,omitempty"`
	MaxPayload       int32                  `json:"max_payload,omitempty"`
	MaxPending       int64                  `json:"max_pending,omitempty"`
	MaxClosedClients int                    `json:"max_closed_clients,omitempty"`
	MaxSubs          int                    `json:"max_subscriptions,omitempty"`
	MaxSubsTokens    uint8                  `json:"max_subscriptions_tokens,omitempty"`
	MaxControlLine   int32                  `json:"max_control_line,omitempty"`
	MaxTracedMsgLen  int                    `json:"max_traced_msg_len,omitempty"`
	MaxPingsOut      int                    `json:"max_pings_out,omitempty"`
	PingInterval     time.Duration          `json:"ping_interval,omitempty"`
	WriteDeadline    time.Duration          `json:"write_deadline,omitempty"`
	NoAuthUser       string                 `json:"no_auth_user,omitempty"`
	Operators        []string               `json:"operators,omitempty"`
	SystemAccount    string                 `json:"system_account,omitempty"`
	Accounts         []*Account             `json:"accounts,omitempty"`
	Authorization    *AuthorizationMap      `json:"authorization,omitempty"`
	FullResolver     *FullAccountResolver   `json:"full_resolver,omitempty"`
	CacheResolver    *CacheAccountResolver  `json:"cache_resolver,omitempty"`
	MemoryResolver   *MemoryAccountResolver `json:"memory_resolver,omitempty"`
	Cluster          *Cluster               `json:"cluster,omitempty"`
	Websocket        *Websocket             `json:"websocket,omitempty"`
	MQTT             *MQTT                  `json:"mqtt,omitempty"`
	JetStream        *JetStream             `json:"jetstream,omitempty"`
	Leafnode         *Leafnode              `json:"leafnode,omitempty"`
	Metrics          *Metrics               `json:"metrics,omitempty"`
}

// SubjectMapping is for mapping published subjects for clients.
type SubjectMapping struct {
	Subject string            `json:"subject"`
	MapDest []*server.MapDest `json:"dest"`
}

// AuthCalloutMap is the configuration for the auth callout.
// It can be used when defining an authorization configuration.
type AuthCalloutMap struct {
	Issuer    string   `json:"issuer,omitempty"`
	AuthUsers []string `json:"auth_users,omitempty"`
	Account   string   `json:"account,omitempty"`
	XKey      string   `json:"xkey,omitempty"`
}

// User is the configuration for a server user.
// It can be used when defining an authorization configuration.
type User struct {
	User                   string              `json:"user,omitempty"`
	Password               string              `json:"password,omitempty"`
	Permissions            *server.Permissions `json:"permissions,omitempty"`
	AllowedConnectionTypes []string            `json:"allowed_connection_types,omitempty"`
}

type AccountLimits = jwt.AccountLimits

// Account is the configuration for a server account.
// It can be used when defining an authorization configuration.
type Account struct {
	Name               string              `json:"name,omitempty"`
	NKey               string              `json:"nkey,omitempty"`
	Users              []User              `json:"users,omitempty"`
	Exports            []string            `json:"exports,omitempty"`
	Imports            []string            `json:"imports,omitempty"`
	JetStream          bool                `json:"jetstream,omitempty"`
	DefaultPermissions *server.Permissions `json:"default_permissions,omitempty"`
	Mappings           []*SubjectMapping   `json:"mappings,omitempty"`
	Limits             *AccountLimits      `json:"limits,omitempty"`
}

// AuthorizationMap block provides authentication configuration
// as well as authorization configuration.
type AuthorizationMap struct {
	Token       string          `json:"token,omitempty"`
	User        string          `json:"user,omitempty"`
	Password    string          `json:"password,omitempty"`
	Users       []User          `json:"users,omitempty"`
	AuthCallout *AuthCalloutMap `json:"auth_callout,omitempty"`
	Timeout     time.Duration   `json:"timeout,omitempty"`
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
	Url         string                `json:"url,omitempty"`
	Urls        []string              `json:"urls,omitempty"`
	Hub         bool                  `json:"hub,omitempty"`
	DenyImports []string              `json:"deny_imports,omitempty"`
	DenyExports []string              `json:"deny_exports,omitempty"`
	NoRandomize bool                  `json:"no_randomize,omitempty"`
	Account     string                `json:"account,omitempty"`
	Credentials string                `json:"credentials,omitempty"`
	Websocket   RemoteWebsocketClient `json:"websocket,omitempty"`
}

// Leafnode is the configuration for the leafnode server
// Leafnode server is disabled by default. It can be enabled
// by providing an empty Leafnode struct (Leafnode{}).
// Default values will be used for missing fields.
type Leafnode struct {
	Host      string   `json:"host,omitempty"`
	Port      int      `json:"port,omitempty"`
	Advertise string   `json:"advertise,omitempty"`
	NoTLS     bool     `json:"no_tls,omitempty"`
	TLS       *TLSMap  `json:"tls,omitempty"`
	Remotes   []Remote `json:"remotes,omitempty"`
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
	ServerLabel    string `json:"server_label,omitempty"`
	ServerUrl      string `json:"server_url,omitempty"`
	Healthz        bool   `json:"healthz,omitempty"`
	Connz          bool   `json:"connz,omitempty"`
	ConnzDetailed  bool   `json:"connz_detailed,omitempty"`
	Subz           bool   `json:"subz,omitempty"`
	Routez         bool   `json:"routez,omitempty"`
	Gatewayz       bool   `json:"gatewayz,omitempty"`
	Leafz          bool   `json:"leafz,omitempty"`
	ReplicatorVarz bool   `json:"replicator_varz,omitempty"`
	JszFilter      string `json:"jsz_filter,omitempty"`
}

// TLSMap is a configuration block for TLSMap servers.
// TLSMap configuration MUST NOT be provided when Let's Encrypt
// certificates are expected to be issued.
type TLSMap struct {
	config           *tls.Config
	Subjects         []string      `json:"subjects,omitempty"`
	AllowNonTLS      bool          `json:"allow_non_tls,omitempty"`
	CertFile         string        `json:"cert_file,omitempty"`
	CertStore        string        `json:"cert_store,omitempty"`
	CertMatch        string        `json:"cert_match,omitempty"`
	CertMatchBy      string        `json:"cert_match_by,omitempty"`
	KeyFile          string        `json:"key_file,omitempty"`
	CaFile           string        `json:"ca_file,omitempty"`
	Verify           bool          `json:"verify,omitempty"`
	Insecure         bool          `json:"insecure,omitempty"`
	Map              bool          `json:"map,omitempty"`
	CheckKnownURLs   bool          `json:"tls_check_known_urls,omitempty"`
	Timeout          time.Duration `json:"timeout,omitempty"`
	RateLimit        int64         `json:"rate_limit,omitempty"`
	Ciphers          []string      `json:"ciphers,omitempty"`
	CurvePreferences []string      `json:"curve_preferences,omitempty"`
	PinnedCerts      []string      `json:"pinned_certs,omitempty"`
}

func (m *TLSMap) IsManaged() bool {
	return m.config != nil || m.Subjects != nil
}

func (m *TLSMap) SetConfigOverride(config *tls.Config) {
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
	m.config = cfg
}
