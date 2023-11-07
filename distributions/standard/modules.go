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
	_ "github.com/quara-dev/beyond/modules/docker/app"
	// NATS modules
	_ "github.com/quara-dev/beyond/modules/nats/app"
	_ "github.com/quara-dev/beyond/modules/nats/auth/callout"
	_ "github.com/quara-dev/beyond/modules/nats/auth/policies"
	_ "github.com/quara-dev/beyond/modules/nats/connectors"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/components/mongo"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/components/nats"

	// OAuth2 modules
	_ "github.com/quara-dev/beyond/modules/oauth2/app"
	_ "github.com/quara-dev/beyond/modules/oauth2/http_handler"
	_ "github.com/quara-dev/beyond/modules/oauth2/stores/jetstream"
	_ "github.com/quara-dev/beyond/modules/oauth2/stores/redis"

	// Telemetry modules
	_ "github.com/quara-dev/beyond/modules/otelcol/app"
	// Python modules
	_ "github.com/quara-dev/beyond/modules/python/app"
	// Secrets modules
	_ "github.com/quara-dev/beyond/modules/secrets/app"
	_ "github.com/quara-dev/beyond/modules/secrets/automation/handlers"
	_ "github.com/quara-dev/beyond/modules/secrets/automation/triggers"
	_ "github.com/quara-dev/beyond/modules/secrets/stores"
)
