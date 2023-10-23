// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package standard

import (
	// plug in Caddy modules here
	_ "github.com/quara-dev/beyond"
	// DNS modules
	_ "github.com/quara-dev/beyond/modules/dns/azure"
	_ "github.com/quara-dev/beyond/modules/dns/digitalocean"

	// Docker modules
	_ "github.com/quara-dev/beyond/modules/docker/dockerapp"
	// NATS modules
	_ "github.com/quara-dev/beyond/modules/nats/auth_callout/allow"
	_ "github.com/quara-dev/beyond/modules/nats/auth_callout/deny"
	_ "github.com/quara-dev/beyond/modules/nats/auth_callout/oauth2"
	_ "github.com/quara-dev/beyond/modules/nats/natsapp"

	// OAuth2 modules
	_ "github.com/quara-dev/beyond/modules/oauth2/http_handler"
	_ "github.com/quara-dev/beyond/modules/oauth2/oauth2app"
	_ "github.com/quara-dev/beyond/modules/oauth2/session_store"
	_ "github.com/quara-dev/beyond/modules/oauth2/session_store/jetstream"

	// Telemetry modules
	_ "github.com/quara-dev/beyond/modules/otelcol"
	// Secrets modules
	_ "github.com/quara-dev/beyond/modules/secrets/secretsapp"
	_ "github.com/quara-dev/beyond/modules/secrets/stores/azure"
	_ "github.com/quara-dev/beyond/modules/secrets/stores/memory"
)
