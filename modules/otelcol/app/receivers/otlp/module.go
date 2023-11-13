// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

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

func (e *OtlpReceiver) ReplaceAll(repl *caddy.Replacer) error {
	if e.Protocols == nil {
		return nil
	}
	if e.Protocols.GRPC != nil {
		if e.Protocols.GRPC.Endpoint == "" {
			ep, err := repl.ReplaceOrErr(e.Protocols.GRPC.Endpoint, true, true)
			if err != nil {
				return err
			}
			e.Protocols.GRPC.Endpoint = ep
		}
	}
	if e.Protocols.HTTP != nil {
		if e.Protocols.HTTP.TracesURLPath != "" {
			tracesURLPath, err := repl.ReplaceOrErr(e.Protocols.HTTP.TracesURLPath, true, true)
			if err != nil {
				return err
			}
			e.Protocols.HTTP.TracesURLPath = tracesURLPath
		}
		if e.Protocols.HTTP.MetricsURLPath != "" {
			metricsURLPath, err := repl.ReplaceOrErr(e.Protocols.HTTP.MetricsURLPath, true, true)
			if err != nil {
				return err
			}
			e.Protocols.HTTP.MetricsURLPath = metricsURLPath
		}
		if e.Protocols.HTTP.LogsURLPath != "" {
			logsURLPath, err := repl.ReplaceOrErr(e.Protocols.HTTP.LogsURLPath, true, true)
			if err != nil {
				return err
			}
			e.Protocols.HTTP.LogsURLPath = logsURLPath
		}
	}
	return nil
}

// Interface guards
var (
	_ config.Receiver = (*OtlpReceiver)(nil)
)
