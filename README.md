# caddy-nats

> Run `nats-server` as a [caddy app](https://caddyserver.com/docs/extending-caddy#app-modules).

## Example usage

First build the project:

```bash
go build ./cmd/caddy
```

Allow caddy to bind to port 80 and 443:

```bash
sudo setcap cap_net_bind_service=+ep ./caddy
```

Start the example:

```bash
./caddy run -c example.json
```

Connect using APP account:

```bash
nats pub foo bar --user APP --password "not used"
```

Connect using SYS account:

```bash
nats server ls --user SYS --passwod "not used"
```

> Note: The example config uses the [`"always_allow"` auth callout module](https://github.com/charbonnierg/caddy-nats/blob/rewrite/modules/auth_callout/allow.go)


### Caddyfile

No support for caddyfile at the moment.

### JSON file

Checkout the file [example.json](./example.json) to see how to configure an NATS server with TLS certificates managed by caddy and auth callout service running as caddy module.

## Next steps

- Use replacers to avoid writing signing key in config
- Add tests
- Add Caddyfile support
- Add auth callout modules (maybe a module validating ID tokens provided by users in connect options ❔)
