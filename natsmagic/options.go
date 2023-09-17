package natsmagic

import (
	"fmt"
	"net/url"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/prometheus-nats-exporter/exporter"
)

type Websocket struct {
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port,omitempty"`
	Advertise string `json:"advertise,omitempty"`
	NoTLS     bool   `json:"no_tls,omitempty"`
}

type MQTT struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	NoTLS    bool   `json:"no_tls,omitempty"`
	JSDomain string `json:"jetstream_domain,omitempty"`
}

type JetStream struct {
	Domain    string `json:"domain,omitempty"`
	StoreDir  string `json:"store_dir,omitempty"`
	MaxMemory int    `json:"max_memory,omitempty"`
	MaxFile   int    `json:"max_file,omitempty"`
}

type Remote struct {
	Url         string `json:"url,omitempty"`
	Account     string `json:"account,omitempty"`
	Credentials string `json:"credentials,omitempty"`
}

type LeafNode struct {
	Host      string   `json:"host,omitempty"`
	Port      int      `json:"port,omitempty"`
	Advertise string   `json:"advertise,omitempty"`
	NoTLS     bool     `json:"no_tls,omitempty"`
	Remotes   []Remote `json:"remotes,omitempty"`
}

type AccountResolver struct {
	Path   string `json:"path,omitempty"`
	Full   bool   `json:"full,omitempty"`
	Memory bool   `json:"memory,omitempty"`
}

type Metrics struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	BasePath string `json:"base_path,omitempty"`
}

func (m *Metrics) GetUrl() string {
	path := m.BasePath
	if path == "" {
		path = "/metrics"
	}
	return fmt.Sprintf("http://%s:%d%s", m.Host, m.Port, path)
}

type NatsConfig struct {
	ConfigFile      string           `json:"config_file,omitempty"`
	SNI             string           `json:"sni,omitempty"`
	ServerName      string           `json:"server_name,omitempty"`
	ServerTags      []string         `json:"server_tags,omitempty"`
	Host            string           `json:"host,omitempty"`
	Port            int              `json:"port,omitempty"`
	ClientAdvertise string           `json:"client_advertise,omitempty"`
	Debug           bool             `json:"debug,omitempty"`
	Trace           bool             `json:"trace,omitempty"`
	TraceVerbose    bool             `json:"trace_verbose,omitempty"`
	HTTPHost        string           `json:"http_host,omitempty"`
	HTTPPort        int              `json:"http_port,omitempty"`
	HTTPSPort       int              `json:"https_port,omitempty"`
	HTTPBasePath    string           `json:"http_base_path,omitempty"`
	NoLog           bool             `json:"disable_logging,omitempty"`
	NoTLS           bool             `json:"no_tls,omitempty"`
	NoSublistCache  bool             `json:"disable_sublist_cache,omitempty"`
	MaxConn         int              `json:"max_connections,omitempty"`
	MaxPayload      int              `json:"max_payload,omitempty"`
	MaxPending      int              `json:"max_pending,omitempty"`
	MaxSubs         int              `json:"max_subscriptions,omitempty"`
	MaxControlLine  int              `json:"max_control_line,omitempty"`
	PingInterval    time.Duration    `json:"ping_interval,omitempty"`
	WriteDeadline   time.Duration    `json:"write_deadline,omitempty"`
	PingMax         int              `json:"ping_max,omitempty"`
	NoAuthUser      string           `json:"no_auth_user,omitempty"`
	Operator        string           `json:"operator,omitempty"`
	SystemAccount   string           `json:"system_account,omitempty"`
	Resolver        *AccountResolver `json:"resolver,omitempty"`
	ResolverPreload []string         `json:"resolver_preload,omitempty"`
	Websocket       *Websocket       `json:"websocket,omitempty"`
	MQTT            *MQTT            `json:"mqtt,omitempty"`
	JetStream       *JetStream       `json:"jetstream,omitempty"`
	LeafNode        *LeafNode        `json:"leafnode,omitempty"`
	Metrics         *Metrics         `json:"metrics,omitempty"`
}

func (o *NatsConfig) GetServerOptions() *server.Options {
	opts := &server.Options{
		ConfigFile:      o.ConfigFile,
		ServerName:      o.ServerName,
		Tags:            o.ServerTags,
		Host:            o.Host,
		Port:            o.Port,
		ClientAdvertise: o.ClientAdvertise,
		Debug:           o.Debug,
		Trace:           o.Trace,
		TraceVerbose:    o.TraceVerbose,
		HTTPPort:        o.HTTPPort,
		HTTPSPort:       o.HTTPSPort,
		HTTPBasePath:    o.HTTPBasePath,
		NoLog:           o.NoLog,
		NoSublistCache:  o.NoSublistCache,
		MaxConn:         o.MaxConn,
		MaxPayload:      int32(o.MaxPayload),
		MaxPending:      int64(o.MaxPending),
		MaxSubs:         o.MaxSubs,
		MaxControlLine:  int32(o.MaxControlLine),
		PingInterval:    o.PingInterval,
		MaxPingsOut:     o.PingMax,
		WriteDeadline:   o.WriteDeadline,
		NoAuthUser:      o.NoAuthUser,
	}
	if o.Websocket != nil {
		opts.Websocket.Host = o.Websocket.Host
		opts.Websocket.Port = o.Websocket.Port
		opts.Websocket.Advertise = o.Websocket.Advertise
		opts.Websocket.NoTLS = o.Websocket.NoTLS
	}
	if o.MQTT != nil {
		opts.MQTT.Host = o.MQTT.Host
		opts.MQTT.Port = o.MQTT.Port
		opts.MQTT.JsDomain = o.MQTT.JSDomain
	}
	if o.JetStream != nil {
		opts.StoreDir = o.JetStream.StoreDir
		opts.JetStream = true
		opts.JetStreamDomain = o.JetStream.Domain
		opts.JetStreamMaxMemory = int64(o.JetStream.MaxMemory)
		opts.JetStreamMaxStore = int64(o.JetStream.MaxFile)
	}
	if o.LeafNode != nil {
		opts.LeafNode.Host = o.LeafNode.Host
		opts.LeafNode.Port = o.LeafNode.Port
		opts.LeafNode.Advertise = o.LeafNode.Advertise
		opts.LeafNode.Remotes = make([]*server.RemoteLeafOpts, len(o.LeafNode.Remotes))
		for i, remote := range o.LeafNode.Remotes {
			remoteUrl, err := url.Parse(remote.Url)
			if err != nil {
				panic(err)
			}
			opts.LeafNode.Remotes[i] = &server.RemoteLeafOpts{
				URLs:         []*url.URL{remoteUrl},
				LocalAccount: remote.Account,
				Credentials:  remote.Credentials,
			}
		}
	}
	if o.Operator != "" {
		claims, err := jwt.DecodeOperatorClaims(o.Operator)
		if err != nil {
			panic(err)
		}
		opts.TrustedOperators = []*jwt.OperatorClaims{claims}
	}
	if o.SystemAccount != "" {
		opts.SystemAccount = o.SystemAccount
	}
	if o.Resolver != nil {
		if o.Resolver.Memory {
			opts.AccountResolver = &server.MemAccResolver{}
		} else if o.Resolver.Full {
			res, err := server.NewDirAccResolver(
				o.Resolver.Path,
				int64(1000),
				time.Duration(2)*time.Minute,
				server.NoDelete,
			)
			if err != nil {
				panic(err)
			}
			opts.AccountResolver = res
		} else {
			res, err := server.NewCacheDirAccResolver(
				o.Resolver.Path,
				int64(1000),
				time.Duration(2)*time.Minute,
			)
			if err != nil {
				panic(err)
			}
			opts.AccountResolver = res
		}
	}
	return opts
}

func (o *NatsConfig) GetMetricsOptions() *exporter.NATSExporterOptions {
	if o.Metrics == nil {
		return nil
	}
	path := o.Metrics.BasePath
	if path == "" {
		path = "/metrics"
	}
	serverUrl := ""
	if o.HTTPPort != 0 {
		serverUrl = fmt.Sprintf("http://localhost:%d", o.HTTPPort)
	} else if o.HTTPSPort != 0 {
		serverUrl = fmt.Sprintf("https://%s:%d", o.SNI, o.HTTPSPort)
	}
	if o.Metrics.Host == "" {
		o.Metrics.Host = "127.0.0.1"
	}
	return &exporter.NATSExporterOptions{
		NATSServerURL:    serverUrl,
		UseServerName:    true,
		ListenAddress:    o.Metrics.Host,
		ListenPort:       o.Metrics.Port,
		ScrapePath:       path,
		GetHealthz:       true,
		GetConnz:         true,
		GetConnzDetailed: true,
		GetVarz:          true,
		GetSubz:          true,
		GetRoutez:        true,
		GetGatewayz:      true,
		GetLeafz:         true,
	}
}
