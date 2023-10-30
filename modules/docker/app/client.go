package app

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type NetworkDefinition struct {
	Name   string
	Config *types.NetworkCreate
}

type ContainerDefinition struct {
	Name             string
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

type NetworkSpec struct {
	ID         string
	Definition *NetworkDefinition
}

type ContainerSpec struct {
	ID         string
	Definition *ContainerDefinition
}

type DockerClient struct {
	client *client.Client
	ctx    context.Context
}

type TlsConfig struct {
	CAFile   string `json:"ca_file,omitempty"`
	CertFile string `json:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
}

type ClientOptions struct {
	Host string     `json:"host,omitempty"`
	Tls  *TlsConfig `json:"tls,omitempty"`
}

func (o *ClientOptions) Opts() []client.Opt {
	opts := []client.Opt{}
	if o.Host != "" {
		opts = append(opts, client.WithHost(o.Host))
	}
	if o.Tls != nil {
		opts = append(opts, client.WithTLSClientConfig(o.Tls.CAFile, o.Tls.CertFile, o.Tls.KeyFile))
	}
	return opts
}

func (c *DockerClient) networkExists(definition *NetworkDefinition) (*NetworkSpec, error) {
	response, err := c.client.NetworkInspect(c.ctx, definition.Name, types.NetworkInspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if response.Driver != definition.Config.Driver {
		return nil, errors.New("network driver mismatch")
	}
	return &NetworkSpec{
		ID:         response.ID,
		Definition: definition,
	}, nil
}

func (c *DockerClient) ProvisionNetwork(definition *NetworkDefinition) (*NetworkSpec, error) {
	existing, err := c.networkExists(definition)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	if definition.Config == nil {
		return nil, errors.New("network config is nil")
	}
	response, err := c.client.NetworkCreate(
		c.ctx,
		definition.Name,
		*definition.Config,
	)
	if err != nil {
		return nil, err
	}
	nid := response.ID
	return &NetworkSpec{
		ID:         nid,
		Definition: definition,
	}, nil
}

func (c *DockerClient) RunContainer(definition *ContainerDefinition) (*ContainerSpec, error) {
	response, err := c.client.ContainerCreate(
		c.ctx,
		definition.Config,
		definition.HostConfig,
		definition.NetworkingConfig,
		nil,
		definition.Name,
	)
	if err != nil {
		return nil, err
	}
	cid := response.ID
	if err := c.client.ContainerStart(c.ctx, cid, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	return &ContainerSpec{
		ID:         cid,
		Definition: definition,
	}, nil
}

func (c *DockerClient) RemoveContainer(cid string) error {
	return c.client.ContainerRemove(c.ctx, cid, types.ContainerRemoveOptions{Force: true})
}

func NewDockerClient(ctx context.Context, opts *ClientOptions) (*DockerClient, error) {
	clientOpts := []client.Opt{}
	if opts != nil {
		clientOpts = append(clientOpts, opts.Opts()...)
	}
	cli, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		client: cli,
		ctx:    ctx,
	}, nil
}
