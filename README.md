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

### Caddyfile

No support for caddyfile at the moment.

### JSON file

Checkout the file [example.json](./example.json) to see how to configure an NATS server with TLS certificates managed by caddy and auth callout service running as caddy module.

## Next steps

- Add tests
- Add Caddyfile support
- Add auth callout modules (maybe a module validating ID tokens provided by users in connect options ❔)
