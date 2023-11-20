// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package standard

import (
	// plug in Caddy modules here
	_ "github.com/quara-dev/beyond"
	_ "github.com/quara-dev/beyond/modules/caddynats"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/azure/eventgridreceiver"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/azure/eventgridwriter"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/mongo/changestreamreader"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/mongo/changestreamwriter"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/nats/consumerreader"
	_ "github.com/quara-dev/beyond/modules/caddynats/connectors/nats/streamwriter"
	_ "github.com/quara-dev/beyond/modules/caddynats/natsauth"
	_ "github.com/quara-dev/beyond/modules/caddynats/natsauth/callout/allow"
	_ "github.com/quara-dev/beyond/modules/caddynats/natsauth/callout/deny"
	_ "github.com/quara-dev/beyond/modules/caddynats/natsauth/callout/oauth2"
	_ "github.com/quara-dev/beyond/modules/caddynats/natsauth/matchers"
	_ "github.com/quara-dev/beyond/modules/caddynats/natscmd"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/jetstream_fs"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/jetstream_get_msg"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/jetstream_kv_get"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/jetstream_kv_put"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/jetstream_publish"
	_ "github.com/quara-dev/beyond/modules/caddynats/natshttp/nats_publish"

	// Logging modules
	_ "github.com/quara-dev/beyond/modules/logs"
	// DNS modules
	_ "github.com/quara-dev/beyond/modules/dns/azure"
	_ "github.com/quara-dev/beyond/modules/dns/digitalocean"

	// Docker modules
	_ "github.com/quara-dev/beyond/modules/docker/app"
	_ "github.com/quara-dev/beyond/modules/docker/nats_matcher"

	_ "github.com/quara-dev/beyond/modules/config/cmd"

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
