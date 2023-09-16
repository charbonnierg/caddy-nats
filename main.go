package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// plug in Caddy modules here
	_ "github.com/caddy-dns/digitalocean"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/charbonnierg/nats-magic/natsmagic"
)

func main() {
	caddycmd.Main()
}
