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
	_ "github.com/quara-dev/beyond/modules/nats/auth/matchers"

	_ "github.com/quara-dev/beyond/modules/config/cmd"
	// Connectors modules
	_ "github.com/quara-dev/beyond/modules/nats/cmd"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/mongo/changestreamexporter"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/mongo/changestreamreceiver"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/mongo/queryservice"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/nats/consumerreceiver"
	_ "github.com/quara-dev/beyond/modules/nats/connectors/nats/streamexporter"
	_ "github.com/quara-dev/beyond/modules/nats/jetstream_fs"
	_ "github.com/quara-dev/beyond/modules/nats/jetstream_publish"
	_ "github.com/quara-dev/beyond/modules/nats/publish"

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
