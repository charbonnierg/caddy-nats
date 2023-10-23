// SPDX-License-Identifier: Apache-2.0

package standard

import (
	// plug in Caddy modules here
	_ "github.com/quara-dev/beyond"
	_ "github.com/quara-dev/beyond/modules/dns/azure"
	_ "github.com/quara-dev/beyond/modules/dns/digitalocean"
	_ "github.com/quara-dev/beyond/modules/docker/dockerapp"
	_ "github.com/quara-dev/beyond/modules/nats/auth_callout"
	_ "github.com/quara-dev/beyond/modules/nats/auth_callout/oauth2"
	_ "github.com/quara-dev/beyond/modules/nats/natsapp"
	_ "github.com/quara-dev/beyond/modules/oauth2/http_handler"
	_ "github.com/quara-dev/beyond/modules/oauth2/oauth2app"
	_ "github.com/quara-dev/beyond/modules/oauth2/session_store"
	_ "github.com/quara-dev/beyond/modules/oauth2/session_store/jetstream"
	_ "github.com/quara-dev/beyond/modules/otelcol"
	_ "github.com/quara-dev/beyond/modules/secrets/secretsapp"
	_ "github.com/quara-dev/beyond/modules/secrets/stores/azure"
	_ "github.com/quara-dev/beyond/modules/secrets/stores/memory"
)
