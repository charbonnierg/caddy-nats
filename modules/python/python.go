// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package python

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
)

type App interface {
	beyond.App
	Replacer() *caddy.Replacer
}
