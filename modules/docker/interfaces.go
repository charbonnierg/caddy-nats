// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// App is the interface implemented by the docker caddy app module.
type App interface{}

type Container interface {
	ContainerConfig(repl *caddy.Replacer) (*container.Config, error)
	ContainerHostConfig(repl *caddy.Replacer) (*container.HostConfig, error)
	ContainerNetworkConfig(repl *caddy.Replacer) (*network.NetworkingConfig, error)
}

type Network interface {
	NetworkConfig(repl *caddy.Replacer) (*types.NetworkCreate, error)
}
