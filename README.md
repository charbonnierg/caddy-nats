# caddy-nats (EXPERIMENTAL)

> Run `nats-server` as a [caddy app](https://caddyserver.com/docs/extending-caddy#app-modules) with experimental oauth2 authentication.

## Example usage

- First build the project:

```bash
go build ./cmd/caddy
```

- Allow caddy to bind to port 80 and 443:

```bash
sudo setcap cap_net_bind_service=+ep ./caddy
```

- Update the  `apps.oauth2` section of the example config. Configuration supports almost all oauth2-proxy options (check the file ./oauthproxy/json_options.go)

  ⚠ example config won't run without change, because it uses Azure Provider with fake data.

Start the example:

```bash
./caddy run -c example.json
```

- Visit `https://localhost`. You should be redirected to configured OAuth2 provider to authenticate. Once authentication is succesfull, you should be redirected back to `https://localhost` and see metrics displayed in the page.

- Open Developer Tools and connect to NATS using oauth2 auth callout:

```javascript
const nats = await import("https://cdn.jsdelivr.net/npm/nats.ws@1.18.0/esm/nats.js")

nc = await nats.connect(
	{"servers": "wss://localhost:10443", "user": "APP", "pass": document.cookie}
)

await nc.publish("test")
```

- Checkout server logs to see that user was authorized by auth callout module and message was successfully published under account `APP`


- Now open a terminal and try to connect using NATS cli:

```bash
nats pub foo bar
```

  An authorization error should be returned, because no username is provided.

- Let's try to connect using SYS account:

```bash
nats server ls --username SYS --password "not used"
```

  > Note: password option is provided because NATS CLI does not send username if no password is provided. However, code using SDK from supported languages can send a connect request with username only.

  The command should succeed, because the matcher for SYS username is configured to handle auth request with `always_allow` handler.

### Caddyfile

No support for caddyfile at the moment.

### JSON file

Checkout the file [example.json](./example.json) to see how to configure an NATS server with TLS certificates managed by caddy and auth callout service running as caddy module.

## Next steps

- Use replacers to avoid writing signing key in config
- Add tests
- Add Caddyfile support
- Add auth callout modules (maybe a module validating ID tokens provided by users in connect options ❔)
