package config

import (
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Filter represent a filter that can be passed to Docker Swarm to reduce the
// amount of data received.
type Filter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// DockerSDConfig is the configuration for Docker (non-swarm) based service discovery.
type DockerSDConfig struct {
	HTTPClientConfig   config.HTTPClientConfig `json:",inline"`
	Host               string                  `json:"host"`
	Port               int                     `json:"port"`
	Filters            []*Filter               `json:"filters"`
	HostNetworkingHost string                  `json:"host_networking_host"`

	RefreshInterval model.Duration `json:"refresh_interval"`
}

// DockerSwarmSDConfig is the configuration for Docker Swarm based service discovery.
type DockerSwarmSDConfig struct {
	HTTPClientConfig config.HTTPClientConfig `json:",inline"`

	Host    string    `json:"host"`
	Role    string    `json:"role"`
	Port    int       `json:"port"`
	Filters []*Filter `json:"filters"`

	RefreshInterval model.Duration `json:"refresh_interval"`
}
