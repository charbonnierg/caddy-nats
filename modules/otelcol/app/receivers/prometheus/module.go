// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	pconfig "github.com/quara-dev/beyond/modules/otelcol/app/receivers/prometheus/config"
)

func init() {
	caddy.RegisterModule(PrometheusReceiver{})
}

type PrometheusReceiver struct {
	Config *pconfig.Config `json:"config,omitempty"`
}

func (PrometheusReceiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.receivers.prometheus",
		New: func() caddy.Module { return new(PrometheusReceiver) },
	}
}

func (e *PrometheusReceiver) ReplaceAll(repl *caddy.Replacer) error {
	return e.Config.ReplaceAll(repl)
}

// Interface guards
var (
	_ config.Receiver = (*PrometheusReceiver)(nil)
)
