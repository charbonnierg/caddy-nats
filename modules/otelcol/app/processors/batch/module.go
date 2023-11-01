package batch

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
)

func init() {
	caddy.RegisterModule(BatchProcessor{})
}

type BatchProcessor struct{}

func (BatchProcessor) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.processors.batch",
		New: func() caddy.Module { return new(BatchProcessor) },
	}
}

// Interface guards
var (
	_ config.Processor = (*BatchProcessor)(nil)
)
