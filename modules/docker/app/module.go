// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	client      *DockerClient
	specs       []*ContainerSpec
	definitions []*ContainerDefinition
	ctx         caddy.Context
	logger      *zap.Logger
	Client      *ClientOptions        `json:"client,omitempty"`
	Containers  map[string]*Container `json:"containers,omitempty"`
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
	a.definitions = []*ContainerDefinition{}
	// This will load the beyond module and register the "secrets" app within beyond module
	b, err := beyond.Register(ctx, a)
	if err != nil {
		return err
	}
	// At this point we can use the beyond module to load other apps
	// Let's load the secret app
	unm, err := b.LoadApp(secrets.NS)
	if err != nil {
		return fmt.Errorf("failed to load secrets app: %v", err)
	}
	secrets := unm.(secrets.App)
	repl := caddy.NewReplacer()
	secrets.AddSecretsReplacerVars(repl)

	client, err := NewDockerClient(a.ctx, a.Client)
	if err != nil {
		return err
	}
	a.client = client
	for name, c := range a.Containers {
		cfg, err := c.ContainerConfig(repl)
		if err != nil {
			return err
		}
		hostCfg, err := c.ContainerHostConfig(repl)
		if err != nil {
			return err
		}
		a.definitions = append(a.definitions, &ContainerDefinition{
			Name:       name,
			Config:     cfg,
			HostConfig: hostCfg,
		})
	}
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Starting docker app")
	for _, d := range a.definitions {
		a.logger.Info("Starting container", zap.String("name", d.Name), zap.String("image", d.Config.Image))
		spec, err := a.client.RunContainer(d)
		if err != nil {
			return err
		}
		a.specs = append(a.specs, spec)
	}
	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping docker app")
	for _, s := range a.specs {
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
