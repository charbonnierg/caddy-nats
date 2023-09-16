package natsmagic

import (
	"net/url"
	"time"

	"github.com/nats-io/nats-server/v2/server"
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

type Options struct {
	ConfigFile      string        `json:"config_file,omitempty"`
	ServerName      string        `json:"server_name,omitempty"`
	ServerTags      []string      `json:"server_tags,omitempty"`
	Host            string        `json:"host,omitempty"`
	Port            int           `json:"port,omitempty"`
	ClientAdvertise string        `json:"client_advertise,omitempty"`
	Debug           bool          `json:"debug,omitempty"`
	Trace           bool          `json:"trace,omitempty"`
	TraceVerbose    bool          `json:"trace_verbose,omitempty"`
	HTTPPort        int           `json:"http_port,omitempty"`
	HTTPSPort       int           `json:"https_port,omitempty"`
	HTTPBasePath    string        `json:"http_base_path,omitempty"`
	NoLog           bool          `json:"disable_logging,omitempty"`
	NoSublistCache  bool          `json:"disable_sublist_cache,omitempty"`
	MaxConn         int           `json:"max_connections,omitempty"`
	MaxPayload      int           `json:"max_payload,omitempty"`
	MaxPending      int           `json:"max_pending,omitempty"`
	MaxSubs         int           `json:"max_subscriptions,omitempty"`
	MaxControlLine  int           `json:"max_control_line,omitempty"`
	PingInterval    time.Duration `json:"ping_interval,omitempty"`
	WriteDeadline   time.Duration `json:"write_deadline,omitempty"`
	PingMax         int           `json:"ping_max,omitempty"`
	NoAuthUser      string        `json:"no_auth_user,omitempty"`
	Websocket       *Websocket    `json:"websocket,omitempty"`
	MQTT            *MQTT         `json:"mqtt,omitempty"`
	JetStream       *JetStream    `json:"jetstream,omitempty"`
	LeafNode        *LeafNode     `json:"leafnode,omitempty"`
}

func (o *Options) GetServerOptions() *server.Options {
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
	return opts
}
