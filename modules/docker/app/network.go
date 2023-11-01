package app

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// IPAM represents IP Address Management
type IPAM struct {
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"` // Per network IPAM driver options
	Config  []*IPAMConfig     `json:"config,omitempty"`  // List of IPAM configuration options, specified as a map: {"Subnet": <CIDR>, "IPRange": <CIDR>, "Gateway": <IP address>, "AuxiliaryAddresses": <map[string]<IP address>>}
}

// IPAMConfig represents IPAM configurations
type IPAMConfig struct {
	Subnet     string            `json:"subnet,omitempty"`
	IPRange    string            `json:"ip_range,omitempty"`
	Gateway    string            `json:"gateway,omitempty"`
	AuxAddress map[string]string `json:"aux_addresses,omitempty"`
}

// Network represents a docker network
type Network struct {
	CheckDuplicate bool              `json:"check_duplicate,omitempty"`
	Driver         string            `json:"driver,omitempty"`
	Scope          string            `json:"scope,omitempty"`
	EnableIPv6     bool              `json:"enable_ipv6,omitempty"`
	IPAM           *IPAM             `json:"ipam,omitempty"`
	Internal       bool              `json:"internal,omitempty"`
	Attachable     bool              `json:"attachable,omitempty"`
	Ingress        bool              `json:"ingress,omitempty"`
	ConfigOnly     bool              `json:"config_only,omitempty"`
	Options        map[string]string `json:"options,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
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
