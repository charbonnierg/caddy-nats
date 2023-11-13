// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package zpages

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
)

func init() {
	caddy.RegisterModule(ZpagesExtension{})
}

type ZpagesExtension struct {
}

func (ZpagesExtension) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.extensions.zpages",
		New: func() caddy.Module { return new(ZpagesExtension) },
	}
}

func (e *ZpagesExtension) ReplaceAll(repl *caddy.Replacer) error {
	return nil
}

// Interface guards
var (
	_ config.Extension = (*ZpagesExtension)(nil)
)
