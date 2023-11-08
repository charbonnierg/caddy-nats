package client

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
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

func (s *ServiceDefinition) Start(nc *nats.Conn) (micro.Service, error) {
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
