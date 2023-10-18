// SPDX-License-Identifier: Apache-2.0

package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	// plug in Caddy modules here
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/charbonnierg/beyond"
	_ "github.com/charbonnierg/beyond/modules/docker"
	_ "github.com/charbonnierg/beyond/modules/nats"
	_ "github.com/charbonnierg/beyond/modules/nats/auth_callout"
	_ "github.com/charbonnierg/beyond/modules/nats/auth_callout/oauth2"
	_ "github.com/charbonnierg/beyond/modules/oauth2"
	_ "github.com/charbonnierg/beyond/modules/oauth2/http_handler"
	_ "github.com/charbonnierg/beyond/modules/oauth2/session_store"
	_ "github.com/charbonnierg/beyond/modules/oauth2/session_store/jetstream"
	_ "github.com/charbonnierg/beyond/modules/secrets"
)

func main() {
	caddycmd.Main()
}
