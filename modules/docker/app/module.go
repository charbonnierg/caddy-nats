// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
	httpcaddyfile.RegisterGlobalOption("docker", parseGlobalOption)
}

type App struct {
	client          *DockerClient
	containersSpecs []*ContainerSpec
	networksSpecs   []*NetworkSpec
	containers      []*ContainerDefinition
	networks        []*NetworkDefinition
	ctx             caddy.Context
	logger          *zap.Logger
	ClientOptions   *ClientOptions        `json:"client,omitempty"`
	Containers      map[string]*Container `json:"containers,omitempty"`
	Networks        map[string]*Network   `json:"networks,omitempty"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "docker",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.containers = []*ContainerDefinition{}
	a.networks = []*NetworkDefinition{}
	repl := caddy.NewReplacer()
	if err := secrets.UpdateReplacer(ctx, repl); err != nil {
		return err
	}
	if a.ClientOptions != nil {
		host, err := repl.ReplaceOrErr(a.ClientOptions.Host, true, true)
		if err != nil {
			return fmt.Errorf("failed to replace host: %w", err)
		}
		a.ClientOptions.Host = host
	}
	client, err := NewDockerClient(a.ctx, a.ClientOptions)
	if err != nil {
		return err
	}
	a.client = client
	for name, n := range a.Networks {
		cfg, err := n.NetworkConfig(repl)
		if err != nil {
			return err
		}
		a.networks = append(a.networks, &NetworkDefinition{
			Name:   name,
			Config: cfg,
		})
	}
	for name, c := range a.Containers {
		cfg, err := c.ContainerConfig(repl)
		if err != nil {
			return err
		}
		hostCfg, err := c.ContainerHostConfig(repl)
		if err != nil {
			return err
		}
		networkCfg, err := c.ContainerNetworkConfig(repl)
		if err != nil {
			return err
		}
		a.containers = append(a.containers, &ContainerDefinition{
			Name:             name,
			Config:           cfg,
			HostConfig:       hostCfg,
			NetworkingConfig: networkCfg,
		})
	}
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Starting docker app")
	for _, n := range a.networks {
		a.logger.Info("Provisioning network", zap.String("name", n.Name))
		spec, err := a.client.ProvisionNetwork(n)
		if err != nil {
			return err
		}
		a.networksSpecs = append(a.networksSpecs, spec)
	}
	for _, d := range a.containers {
		a.logger.Info("Starting container", zap.String("name", d.Name), zap.String("image", d.Config.Image))
		spec, err := a.client.ProvisionContainer(d)
		if err != nil {
			return err
		}
		a.containersSpecs = append(a.containersSpecs, spec)
	}
	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping docker app")
	for _, s := range a.containersSpecs {
		a.logger.Info("Stopping container", zap.String("name", s.Definition.Name), zap.String("image", s.Definition.Config.Image))
		if err := a.client.RemoveContainer(s.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Validate() error {
	return nil
}

var (
	// Make sure app implements the beyond.App interface
	_ beyond.App = (*App)(nil)
	// Only methods exposed by the interfaces.SecretApp interface will be accessible
	// to other apps.
	_ docker.App = (*App)(nil)
)
