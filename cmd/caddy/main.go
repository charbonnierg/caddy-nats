// SPDX-License-Identifier: Apache-2.0

package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	// plug in Caddy modules here
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/charbonnierg/beyond"
	_ "github.com/charbonnierg/beyond/modules/docker/dockerapp"
	_ "github.com/charbonnierg/beyond/modules/nats/auth_callout"
	_ "github.com/charbonnierg/beyond/modules/nats/auth_callout/oauth2"
	_ "github.com/charbonnierg/beyond/modules/nats/natsapp"
	_ "github.com/charbonnierg/beyond/modules/oauth2/http_handler"
	_ "github.com/charbonnierg/beyond/modules/oauth2/oauth2app"
	_ "github.com/charbonnierg/beyond/modules/oauth2/session_store"
	_ "github.com/charbonnierg/beyond/modules/oauth2/session_store/jetstream"
	_ "github.com/charbonnierg/beyond/modules/secrets/secretsapp"
)

func main() {
	caddycmd.Main()
}
