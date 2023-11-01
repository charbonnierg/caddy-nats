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

// func (p *PrometheusReceiver) MarshalJSON() ([]byte, error) {
// 	// for _, cfg := range p.Config.ScrapeConfigs {
// 	// 	cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &promconfig.DefaultHTTPClientConfig)
// 	// }
// 	return json.Marshal(map[string]*pconfig.Config{"config": p.Config})
// }

// Interface guards
var (
	_ config.Receiver = (*PrometheusReceiver)(nil)
)
