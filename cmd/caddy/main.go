// SPDX-License-Identifier: Apache-2.0

package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	// plug in Caddy modules here
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/charbonnierg/caddy-nats/modules"
	_ "github.com/charbonnierg/caddy-nats/modules/auth_callout"
	_ "github.com/charbonnierg/caddy-nats/modules/auth_callout/oauth2"
	_ "github.com/charbonnierg/caddy-nats/oauthproxy"
	_ "github.com/charbonnierg/caddy-nats/oauthproxy/http_handler"
)

func main() {
	caddycmd.Main()
}
