package otlp

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
)

func init() {
	caddy.RegisterModule(OtlpReceiver{})
}

type HTTPConfig struct {
	*settings.HTTPServerSettings

	// The URL path to receive traces on. If omitted "/v1/traces" will be used.
	TracesURLPath string `json:"traces_url_path,omitempty"`

	// The URL path to receive metrics on. If omitted "/v1/metrics" will be used.
	MetricsURLPath string `json:"metrics_url_path,omitempty"`

	// The URL path to receive logs on. If omitted "/v1/logs" will be used.
	LogsURLPath string `json:"logs_url_path,omitempty"`
}

type Protocols struct {
	GRPC *settings.GRPCServerSettings `json:"grpc,omitempty"`
	HTTP *HTTPConfig                  `json:"http,omitempty"`
}

type OtlpReceiver struct {
	Protocols *Protocols `json:"protocols,omitempty"`
}

func (OtlpReceiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.receivers.otlp",
		New: func() caddy.Module { return new(OtlpReceiver) },
	}
}

// Interface guards
var (
	_ config.Receiver = (*OtlpReceiver)(nil)
)
