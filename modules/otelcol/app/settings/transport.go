package settings

import "time"

type CORSSettings struct {
	// AllowedOrigins sets the allowed values of the Origin header for
	// HTTP/JSON requests to an OTLP receiver. An origin may contain a
	// wildcard (*) to replace 0 or more characters (e.g.,
	// "http://*.domain.com", or "*" to allow any origin).
	AllowedOrigins []string `json:"allowed_origins,omitempty"`

	// AllowedHeaders sets what headers will be allowed in CORS requests.
	// The Accept, Accept-Language, Content-Type, and Content-Language
	// headers are implicitly allowed. If no headers are listed,
	// X-Requested-With will also be accepted by default. Include "*" to
	// allow any request header.
	AllowedHeaders []string `json:"allowed_headers,omitempty"`

	// MaxAge sets the value of the Access-Control-Max-Age response header.
	// Set it to the number of seconds that browsers should cache a CORS
	// preflight response for.
	MaxAge int `json:"max_age,omitempty"`
}

// HTTPServerSettings defines settings for creating an HTTP server.
type HTTPServerSettings struct {
	// Endpoint configures the listening address for the server.
	Endpoint string `json:"endpoint,omitempty"`

	// TLSSetting struct exposes TLS client configuration.
	TLSSetting *TLSServerSetting `json:"tls,omitempty"`

	// CORS configures the server for HTTP cross-origin resource sharing (CORS).
	CORS *CORSSettings `json:"cors,omitempty"`

	// Auth for this receiver
	Auth *Authentication `json:"auth,omitempty"`

	// MaxRequestBodySize sets the maximum request body size in bytes
	MaxRequestBodySize int64 `json:"max_request_body_size,omitempty"`

	// Additional headers attached to each HTTP response sent to the client.
	// Header values are opaque since they may be sensitive.
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
}

// KeepaliveServerConfig is the configuration for keepalive.
type KeepaliveServerConfig struct {
	ServerParameters  *KeepaliveServerParameters  `json:"server_parameters,omitempty"`
	EnforcementPolicy *KeepaliveEnforcementPolicy `json:"enforcement_policy,omitempty"`
}

// KeepaliveServerParameters allow configuration of the keepalive.ServerParameters.
// The same default values as keepalive.ServerParameters are applicable and get applied by the server.
// See https://godoc.org/google.golang.org/grpc/keepalive#ServerParameters for details.
type KeepaliveServerParameters struct {
	MaxConnectionIdle     time.Duration `json:"max_connection_idle,omitempty"`
	MaxConnectionAge      time.Duration `json:"max_connection_age,omitempty"`
	MaxConnectionAgeGrace time.Duration `json:"max_connection_age_grace,omitempty"`
	Time                  time.Duration `json:"time,omitempty"`
	Timeout               time.Duration `json:"timeout,omitempty"`
}

// KeepaliveEnforcementPolicy allow configuration of the keepalive.EnforcementPolicy.
// The same default values as keepalive.EnforcementPolicy are applicable and get applied by the server.
// See https://godoc.org/google.golang.org/grpc/keepalive#EnforcementPolicy for details.
type KeepaliveEnforcementPolicy struct {
	MinTime             time.Duration `json:"min_time,omitempty"`
	PermitWithoutStream bool          `json:"permit_without_stream,omitempty"`
}

// GRPCServerSettings defines common settings for a gRPC server configuration.
type GRPCServerSettings struct {
	// Endpoint configures the address for this network connection.
	// For TCP and UDP networks, the address has the form "host:port". The host must be a literal IP address,
	// or a host name that can be resolved to IP addresses. The port must be a literal port number or a service name.
	// If the host is a literal IPv6 address it must be enclosed in square brackets, as in "[2001:db8::1]:80" or
	// "[fe80::1%zone]:80". The zone specifies the scope of the literal IPv6 address as defined in RFC 4007.
	Endpoint string `json:"endpoint,omitempty"`

	// Transport to use. Known protocols are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),
	// "udp6" (IPv6-only), "ip", "ip4" (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and "unixpacket".
	Transport string `json:"transport,omitempty"`

	// Configures the protocol to use TLS.
	// The default value is nil, which will cause the protocol to not use TLS.
	TLSSetting *TLSServerSetting `json:"tls,omitempty"`

	// MaxRecvMsgSizeMiB sets the maximum size (in MiB) of messages accepted by the server.
	MaxRecvMsgSizeMiB uint64 `json:"max_recv_msg_size_mib,omitempty"`

	// MaxConcurrentStreams sets the limit on the number of concurrent streams to each ServerTransport.
	// It has effect only for streaming RPCs.
	MaxConcurrentStreams uint32 `json:"max_concurrent_streams,omitempty"`

	// ReadBufferSize for gRPC server. See grpc.ReadBufferSize.
	// (https://godoc.org/google.golang.org/grpc#ReadBufferSize).
	ReadBufferSize int `json:"read_buffer_size,omitempty"`

	// WriteBufferSize for gRPC server. See grpc.WriteBufferSize.
	// (https://godoc.org/google.golang.org/grpc#WriteBufferSize).
	WriteBufferSize int `json:"write_buffer_size,omitempty"`

	// Keepalive anchor for all the settings related to keepalive.
	Keepalive *KeepaliveServerConfig `json:"keepalive,omitempty"`

	// Auth for this receiver
	Auth *Authentication `json:"auth,omitempty"`
}
