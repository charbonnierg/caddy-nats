package auth_callout

import (
	_ "github.com/quara-dev/beyond/modules/nats/auth/callout/allow"
	_ "github.com/quara-dev/beyond/modules/nats/auth/callout/deny"
	_ "github.com/quara-dev/beyond/modules/nats/auth/callout/oauth2"
)
