package natsclient

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

type EndpointDefinition struct {
	// Name of the endpoint.
	Name string `json:"name"`
	// Subject on which the endpoint is registered.
	Subject string `json:"subject"`
	// Metadata annotates the service
	Metadata map[string]string `json:"metadata,omitempty"`
	// QueueGroup can be used to override the default queue group name.
	QueueGroup string `json:"queue_group"`
	// Handler used by the endpoint.
	Handler micro.Handler
}

type ServiceDefinition struct {
	Name string `json:"name,omitempty"`
	// Endpoint is a list of endpoint configuration.
	Endpoints []*EndpointDefinition `json:"endpoints,omitempty"`
	// Version is a SemVer compatible version string.
	Version string `json:"version,omitempty"`
	// Description of the service.
	Description string `json:"description,omitempty"`
	// Metadata annotates the service
	Metadata map[string]string `json:"metadata,omitempty"`
	// QueueGroup can be used to override the default queue group name.
	QueueGroup string `json:"queue_group,omitempty"`
	// StatsHandler is a user-defined custom function.
	// used to calculate additional service stats.
	StatsHandler micro.StatsHandler
	// DoneHandler is invoked when all service subscription are stopped.
	DoneHandler micro.DoneHandler
	// ErrorHandler is invoked on any nats-related service error.
	ErrorHandler micro.ErrHandler
}

func (s *ServiceDefinition) Configure(nc *nats.Conn) (micro.Service, error) {
	if len(s.Endpoints) == 0 {
		return nil, errors.New("no endpoint defined")
	}
	defaultEndpoint := s.Endpoints[0]
	var name = s.Name
	if name == "" {
		name = defaultEndpoint.Name
	}
	cfg := micro.Config{
		Name:        name,
		Version:     s.Version,
		Description: s.Description,
		Metadata:    s.Metadata,
		QueueGroup:  s.QueueGroup,
		Endpoint: &micro.EndpointConfig{
			Subject:    defaultEndpoint.Subject,
			Handler:    defaultEndpoint.Handler,
			Metadata:   defaultEndpoint.Metadata,
			QueueGroup: defaultEndpoint.QueueGroup,
		},
		StatsHandler: s.StatsHandler,
		DoneHandler:  s.DoneHandler,
		ErrorHandler: s.ErrorHandler,
	}
	srv, err := micro.AddService(nc, cfg)
	if err != nil {
		return nil, err
	}
	for _, endpoint := range s.Endpoints[1:] {
		err := srv.AddEndpoint(
			endpoint.Name,
			endpoint.Handler,
			micro.WithEndpointSubject(endpoint.Subject),
			micro.WithEndpointMetadata(endpoint.Metadata),
			micro.WithEndpointQueueGroup(endpoint.QueueGroup),
		)
		if err != nil {
			srv.Stop()
			return nil, err
		}
	}
	return srv, nil
}

type ServiceProvider interface {
	Definition() (*ServiceDefinition, error)
}

type ServiceProviderModule interface {
	caddy.Module
	ServiceProvider
	Provision(ctx caddy.Context) error
}

func LoadRawServiceProvider(d *caddyfile.Dispenser, field string) (json.RawMessage, error) {
	var service string
	if err := parser.ParseString(d, &service); err != nil {
		return nil, err
	}
	unm, err := caddyfile.UnmarshalModule(d, "nats_server.services."+service)
	if err != nil {
		return nil, err
	}
	s, ok := unm.(ServiceProviderModule)
	if !ok {
		return nil, d.Errf("service '%s' invalid type", service)
	}
	if field == "" {
		return caddyconfig.JSON(s, nil), nil
	}
	return caddyconfig.JSONModuleObject(s, field, s.CaddyModule().ID.Name(), nil), nil
}
