package connectors

import "github.com/caddyserver/caddy/v2"

// ConnectorsApp is a Caddy app that manages connectors.
type ConnectorsApp interface {
	caddy.Module
	caddy.Provisioner
	caddy.App
	Context() caddy.Context
}

// InputConnector is a Caddy module that serves as a connector
// to a data source. It reads data from a data source and sends it
// to a stream.
type InputConnector interface {
	caddy.Module
	Provision(app ConnectorsApp) error
	Start() error
	Stop() error
}

// OutputConnector is a Caddy module that serves as a connector
// to a data source. It writes data coming from a stream to a data source.
type OutputConnector interface {
	caddy.Module
	Provision(app ConnectorsApp) error
	Start() error
	Stop() error
}
