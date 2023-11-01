package otlphttp

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
)

func init() {
	caddy.RegisterModule(OtlpHttpExporter{})
}

// HTTPClientSettings defines settings for creating an HTTP client.
type HTTPClientSettings struct {
	// The target URL to send data to (e.g.: http://some.url:9411/v1/traces).
	Endpoint string `json:"endpoint,omitempty"`

	// TLSSetting struct exposes TLS client configuration.
	TLSSetting *settings.TLSClientSetting `json:"tls,omitempty"`

	// ReadBufferSize for HTTP client. See http.Transport.ReadBufferSize.
	ReadBufferSize int `json:"read_buffer_size,omitempty"`

	// WriteBufferSize for HTTP client. See http.Transport.WriteBufferSize.
	WriteBufferSize int `json:"write_buffer_size,omitempty"`

	// Timeout parameter configures `http.Client.Timeout`.
	Timeout time.Duration `json:"timeout,omitempty"`

	// Additional headers attached to each HTTP request sent by the client.
	// Existing header values are overwritten if collision happens.
	// Header values are opaque since they may be sensitive.
	Headers map[string]string `json:"headers,omitempty"`

	// Auth configuration for outgoing HTTP calls.
	Auth *settings.Authentication `json:"auth,omitempty"`

	// The compression key for supported compression types within collector.
	Compression string `json:"compression,omitempty"`

	// MaxIdleConns is used to set a limit to the maximum idle HTTP connections the client can keep open.
	// There's an already set value, and we want to override it only if an explicit value provided
	MaxIdleConns *int `json:"max_idle_conns,omitempty"`

	// MaxIdleConnsPerHost is used to set a limit to the maximum idle HTTP connections the host can keep open.
	// There's an already set value, and we want to override it only if an explicit value provided
	MaxIdleConnsPerHost *int `json:"max_idle_conns_per_host,omitempty"`

	// MaxConnsPerHost limits the total number of connections per host, including connections in the dialing,
	// active, and idle states.
	// There's an already set value, and we want to override it only if an explicit value provided
	MaxConnsPerHost *int `json:"max_conns_per_host,omitempty"`

	// IdleConnTimeout is the maximum amount of time a connection will remain open before closing itself.
	// There's an already set value, and we want to override it only if an explicit value provided
	IdleConnTimeout *time.Duration `json:"idle_conn_timeout,omitempty"`

	// DisableKeepAlives, if true, disables HTTP keep-alives and will only use the connection to the server
	// for a single HTTP request.
	//
	// WARNING: enabling this option can result in significant overhead establishing a new HTTP(S)
	// connection for every request. Before enabling this option please consider whether changes
	// to idle connection settings can achieve your goal.
	DisableKeepAlives bool `json:"disable_keep_alives,omitempty"`
}

type QueueSettings struct {
	// Enabled indicates whether to not enqueue batches before sending to the consumerSender.
	Enabled bool `json:"enabled,omitempty"`
	// NumConsumers is the number of consumers from the queue.
	NumConsumers int `json:"num_consumers,omitempty"`
	// QueueSize is the maximum number of batches allowed in queue at a given time.
	QueueSize int `json:"queue_size,omitempty"`
	// StorageID if not empty, enables the persistent storage and uses the component specified
	// as a storage extension for the persistent queue
	StorageID *string `json:"storage,omitempty"`
}

// RetrySettings defines configuration for retrying batches in case of export failure.
// The current supported strategy is exponential backoff.
type RetrySettings struct {
	// Enabled indicates whether to not retry sending batches in case of export failure.
	Enabled bool `json:"enabled,omitempty"`
	// InitialInterval the time to wait after the first failure before retrying.
	InitialInterval time.Duration `json:"initial_interval,omitempty"`
	// RandomizationFactor is a random factor used to calculate next backoffs
	// Randomized interval = RetryInterval * (1 ± RandomizationFactor)
	RandomizationFactor float64 `json:"randomization_factor,omitempty"`
	// Multiplier is the value multiplied by the backoff interval bounds
	Multiplier float64 `json:"multiplier,omitempty"`
	// MaxInterval is the upper bound on backoff interval. Once this value is reached the delay between
	// consecutive retries will always be `MaxInterval`.
	MaxInterval time.Duration `json:"max_interval,omitempty"`
	// MaxElapsedTime is the maximum amount of time (including retries) spent trying to send a request/batch.
	// Once this value is reached, the data is discarded.
	MaxElapsedTime time.Duration `json:"max_elapsed_time,omitempty"`
}

type OtlpHttpExporter struct {
	HTTPClientSettings
	SendingQueue  *QueueSettings `json:"sending_queue,omitempty"`
	RetrySettings *RetrySettings `json:"retry_on_failure,omitempty"`

	// The URL to send traces to. If omitted the Endpoint + "/v1/traces" will be used.
	TracesEndpoint string `json:"traces_endpoint,omitempty"`

	// The URL to send metrics to. If omitted the Endpoint + "/v1/metrics" will be used.
	MetricsEndpoint string `json:"metrics_endpoint,omitempty"`

	// The URL to send logs to. If omitted the Endpoint + "/v1/logs" will be used.
	LogsEndpoint string `json:"logs_endpoint,omitempty"`
}

func (OtlpHttpExporter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.exporters.otlphttp",
		New: func() caddy.Module { return new(OtlpHttpExporter) },
	}
}

// Interface guards
var (
	_ config.Exporter = (*OtlpHttpExporter)(nil)
)
