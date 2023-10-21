# OTEL Collector Components

The code in this directory is generated using [`opentelemetry collector builder (ocb)`](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder#opentelemetry-collector-builder-ocb).

## Generate code

> The current code was generated using version `v0.87.0`.

- Download `ocb` from GitHub releases: <https://github.com/open-telemetry/opentelemetry-collector/releases/tag/cmd%2Fbuilder%2Fv0.87.0>. For example for Linux users on modern laptop or servers:

```bash
wget -O ocb https://github.com/open-telemetry/opentelemetry-collector/releases/download/cmd%2Fbuilder%2Fv0.87.0/ocb_0.87.0_linux_amd64
chmod +x ocb
```

- Generate code using the [manifest.yml file](./manifest.yml):

```bash
task generate:otelcol
```

- Remove unused files, edit package name within remaining files, and rename `components` function into `Components` function in the [components.go module file](./components.go).

- Tidy up go module:

```bash
task tidy
```