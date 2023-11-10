package natsmatcher

import (
	"errors"
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func init() {
	caddy.RegisterModule(DockerHostMatcher{})
}

type DockerHostMatcher struct {
	app       docker.App
	ctx       caddy.Context
	Container string            `json:"container,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Network   []string          `json:"network,omitempty"`
}

func (DockerHostMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.docker",
		New: func() caddy.Module { return new(DockerHostMatcher) },
	}
}

func (m *DockerHostMatcher) Provision(app nats.App) error {
	m.ctx = app.Context()
	if m.Container != "" && m.Labels != nil {
		return errors.New("container and labels are mutually exclusive")
	}
	if m.Container != "" && m.Network != nil {
		return errors.New("container and network are mutually exclusive")
	}
	if m.Labels != nil && m.Network != nil {
		return errors.New("labels and network are mutually exclusive")
	}
	unm, err := app.LoadBeyondApp("docker")
	if err != nil {
		return err
	}
	dockerApp, ok := unm.(docker.App)
	if !ok {
		return errors.New("failed to load docker app")
	}
	m.app = dockerApp
	return nil
}

func (m *DockerHostMatcher) Match(request *jwt.AuthorizationRequestClaims) bool {
	if m.Container != "" {
		return m.matchByContainerName(request)
	}
	if m.Labels != nil {
		return m.matchByLabels(request)
	}
	if m.Network != nil {
		return m.matchByNetwork(request)
	}
	return false
}

// Match by network: query containers with such network, and check if host ip address matches
// one of the container ip address
func (m *DockerHostMatcher) matchByNetwork(request *jwt.AuthorizationRequestClaims) bool {
	client, err := m.app.Client()
	if err != nil {
		return false
	}
	networks := []filters.KeyValuePair{}
	for _, network := range m.Network {
		networks = append(networks, filters.KeyValuePair{
			Key:   "network",
			Value: network,
		})
	}
	response, err := client.ContainerList(m.ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			networks...,
		),
	})
	if err != nil {
		return false
	}
	for _, container := range response {
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress == request.ClientInformation.Host {
				return true
			}
		}
	}
	return false
}

// Match by labels: query containers with such labels, and check if host ip address matches
// one of the container ip address
func (m *DockerHostMatcher) matchByLabels(request *jwt.AuthorizationRequestClaims) bool {
	client, err := m.app.Client()
	if err != nil {
		return false
	}
	labels := []filters.KeyValuePair{}
	for k, v := range m.Labels {
		labels = append(labels, filters.KeyValuePair{
			Key:   "label",
			Value: fmt.Sprintf("%s=%s", k, v),
		})
	}
	response, err := client.ContainerList(m.ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			labels...,
		),
	})
	if err != nil {
		return false
	}
	for _, container := range response {
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress == request.ClientInformation.Host {
				return true
			}
		}
	}
	return false
}

// Match by container name: query container with such name, and check if host ip address matches
// one of the container ip address
func (m *DockerHostMatcher) matchByContainerName(request *jwt.AuthorizationRequestClaims) bool {
	client, err := m.app.Client()
	if err != nil {
		return false
	}
	container, err := client.ContainerInspect(m.ctx, m.Container)
	if err != nil {
		return false
	}
	if container.NetworkSettings.IPAddress == request.ClientInformation.Host {
		return true
	}
	for _, network := range container.NetworkSettings.Networks {
		if network.IPAddress == request.ClientInformation.Host {
			return true
		}
	}
	return false
}

func (m *DockerHostMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "container":
			if err := parser.ParseString(d, &m.Container); err != nil {
				return err
			}
		case "labels":
			if err := parser.ParseStringMap(d, &m.Labels); err != nil {
				return err
			}
		case "network":
			if err := parser.ParseStringArray(d, &m.Network); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown property '%s'", d.Val())
		}
	}
	return nil
}
