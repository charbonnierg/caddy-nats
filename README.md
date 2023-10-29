# Beyond (EXPERIMENTAL)

> Run `nats-server` as a [caddy app](https://caddyserver.com/docs/extending-caddy#app-modules) with experimental oauth2 authentication.

> Since a few days, this repo became much more than running NATS together with caddy, it also includes:
> - OpenTelemetry Collector
> - DNS providers for DNS-01 challenge
> - Secrets provider
> But this is a work-in-progress, not documented yet.

## Introduction

As of NATS v2.10.0, Auth Callout is an opt-in extension for delegating client authentication and authorization to an application-defined NATS service.

The reference for Auth Callout is in the [ADR-26: NATS Authorization Callouts](https://github.com/nats-io/nats-architecture-and-design/blob/main/adr/ADR-26.md)

As stated in the ADR, Authorization Callout aims to enable an external NATS service to generate NATS authorization and credentials by authenticating connection requests.

[The documentation](https://docs.nats.io/running-a-nats-service/configuration/securing_nats/auth_callout) helps understanding the usecases:

> The motivation for this extension is to support applications using an alternate identity and access management (IAM) backend as the source of truth for managing users/applications/machines credentials and permissions. This could be services that implement standard protocols such as LDAP, SAML, and OAuth, an ad-hoc database, or even a file on disk.

Since this feature is quite new, there isn't much documentation aside from the two links above ([ADR-26](https://github.com/nats-io/nats-architecture-and-design/blob/main/adr/ADR-26.md) and [NATS docs](https://github.com/nats-io/nats-architecture-and-design/blob/main/adr/ADR-26.md)), but we can at least draw a few things:

- Authorization callout addresses the need for external authentication/authorization
- Authorization callout is performed by a service which connects to the NATS server (not the NATS server itself)
- Auth callout service receives an authorization request and must return either an error or a signed authorization response

Assuming that a project wants OAuth2 authenticated web users to connect to NATS using authorization callout, the following components are required:

- a NATS server configured to use authorization callout
- an HTTP server to serve the web application
- an OAuth2 middleware to ensure that HTTP sessions are authenticated and authorized
- an auth callout NATS service connected to the NATS server and verifying session state before issuing user claims

The goal of this repository is to provide a single executable binary, which will act as:
- a NATS server
- a NATS authorization callout service 
- an HTTP server
- an Oauth2 middleware

It means that no additional software or components should be required in order to authenticate, authorize and allow users to connect to NATS.

## Introduction to caddy

[Caddy](https://caddyserver.com/) is ...

## Introduction to oauth2-proxy

OAuth2 is not trivial to implement (even though lots of libraries exist to help developers integrate OAuth2 authorization into their applications). In order to avoid writing an OAuth2 middleware, I decided to reuse the existing project [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy). This project is used by many companies or other open sources projects, has more than 300 contributors, and approximately 7600 stars on GitHub. Because this project is a **server** component, it cannot be integrated directly into an existing Go application. I had to [make a fork to use oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy/compare/master...quara-dev:oauth2-proxy:library_usage) in order to easily create a [caddy module](https://github.com/quara-dev/beyond/blob/rewrite/oauthproxy/app.go). This caddy module is an HTTP middleware which can be used before other caddy modules to authenticate and authorize an HTTP session.

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
./caddy run
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

Checkout the [example Caddyfile](./Caddyfile).

### JSON file

Checkout the [example caddy.json](./example.json) equivalent to [example Caddyfile](./Caddyfile)
## Next steps

- Use replacers to avoid writing signing key in config
- Add tests
- Add Caddyfile support
- Add auth callout modules (maybe a module validating ID tokens provided by users in connect options ❔)
