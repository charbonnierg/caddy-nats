package app

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// IPAM represents IP Address Management
type IPAM struct {
	Driver  string
	Options map[string]string // Per network IPAM driver options
	Config  []*IPAMConfig
}

// IPAMConfig represents IPAM configurations
type IPAMConfig struct {
	Subnet     string            `json:",omitempty"`
	IPRange    string            `json:",omitempty"`
	Gateway    string            `json:",omitempty"`
	AuxAddress map[string]string `json:"AuxiliaryAddresses,omitempty"`
}

// Network represents a docker network
type Network struct {
	CheckDuplicate bool
	Driver         string
	Scope          string
	EnableIPv6     bool
	IPAM           *IPAM
	Internal       bool
	Attachable     bool
	Ingress        bool
	ConfigOnly     bool
	Options        map[string]string
	Labels         map[string]string
}

func (n *Network) ipamConfig() (*network.IPAM, error) {
	if n.IPAM == nil {
		return nil, nil
	}
	ipam := &network.IPAM{
		Driver:  n.IPAM.Driver,
		Options: n.IPAM.Options,
		Config:  make([]network.IPAMConfig, len(n.IPAM.Config)),
	}
	for idx, config := range n.IPAM.Config {
		ipam.Config[idx] = network.IPAMConfig{
			Subnet:     config.Subnet,
			IPRange:    config.IPRange,
			Gateway:    config.Gateway,
			AuxAddress: config.AuxAddress,
		}
	}
	return ipam, nil
}

func (n *Network) NetworkConfig(repl *caddy.Replacer) (*types.NetworkCreate, error) {
	ipam, err := n.ipamConfig()
	if err != nil {
		return nil, err
	}
	network := &types.NetworkCreate{
		CheckDuplicate: n.CheckDuplicate,
		Driver:         n.Driver,
		Scope:          n.Scope,
		EnableIPv6:     n.EnableIPv6,
		IPAM:           ipam,
		Internal:       n.Internal,
		Attachable:     n.Attachable,
		Ingress:        n.Ingress,
		ConfigOnly:     n.ConfigOnly,
		Options:        n.Options,
		Labels:         n.Labels,
	}
	return network, nil
}
