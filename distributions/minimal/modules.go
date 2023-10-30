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

	// OAuth2 modules
	_ "github.com/quara-dev/beyond/modules/oauth2/app"
	_ "github.com/quara-dev/beyond/modules/oauth2/http_handler"
	_ "github.com/quara-dev/beyond/modules/oauth2/stores/jetstream"

	// Secrets modules
	_ "github.com/quara-dev/beyond/modules/secrets/app"
	_ "github.com/quara-dev/beyond/modules/secrets/automation/handlers"
	_ "github.com/quara-dev/beyond/modules/secrets/automation/triggers"
	_ "github.com/quara-dev/beyond/modules/secrets/stores"
)
