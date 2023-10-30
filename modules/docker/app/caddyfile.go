package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	a := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, a)
		if err != nil {
			return nil, err
		}
	}
	err := a.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "docker",
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "client":
				a.Client = fnutils.DefaultIfNil(a.Client, new(ClientOptions))
				if err := parseClientOption(d, a.Client); err != nil {
					return err
				}
			case "network":
				a.Networks = fnutils.DefaultIfEmptyMap(a.Networks, make(map[string]*Network))
				if err := parseNetworkOption(d, a.Networks); err != nil {
					return err
				}
			case "container":
				if a.Containers == nil {
					a.Containers = make(map[string]*Container)
				}
				if err := parseContainerOption(d, a.Containers); err != nil {
					return err
				}
			default:
				return d.Errf("unknown option '%s'", d.Val())
			}
		}
	}
	return nil
}

func parseClientOption(d *caddyfile.Dispenser, opts *ClientOptions) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "host":
			if err := caddyutils.ParseString(d, &opts.Host); err != nil {
				return err
			}
		case "tls":
			opts.Tls = fnutils.DefaultIfNil(opts.Tls, new(TlsConfig))
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "ca_file":
					if err := caddyutils.ParseString(d, &opts.Tls.CAFile); err != nil {
						return err
					}
				case "cert_file":
					if err := caddyutils.ParseString(d, &opts.Tls.CertFile); err != nil {
						return err
					}
				case "key_file":
					if err := caddyutils.ParseString(d, &opts.Tls.KeyFile); err != nil {
						return err
					}
				default:
					return d.Errf("unknown option '%s'", d.Val())
				}
			}
		default:
			return d.Errf("unknown option '%s'", d.Val())
		}
	}
	return nil
}

func parseNetworkOption(d *caddyfile.Dispenser, networks map[string]*Network) error {
	n := new(Network)
	name := ""
	if err := caddyutils.ParseString(d, &name); err != nil {
		return err
	}
	if name == "" {
		return d.Errf("network name is required")
	}
	_, ok := networks[name]
	if ok {
		return d.Errf("network '%s' already defined", name)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "driver":
			if err := caddyutils.ParseString(d, &n.Driver); err != nil {
				return err
			}
		case "scope":
			if err := caddyutils.ParseString(d, &n.Scope); err != nil {
				return err
			}
		case "enable_ipv6":
			if err := caddyutils.ParseBool(d, &n.EnableIPv6); err != nil {
				return err
			}
		case "internal":
			if err := caddyutils.ParseBool(d, &n.Internal); err != nil {
				return err
			}
		case "attachable":
			if err := caddyutils.ParseBool(d, &n.Attachable); err != nil {
				return err
			}
		case "ingress":
			if err := caddyutils.ParseBool(d, &n.Ingress); err != nil {
				return err
			}
		case "config_only":
			if err := caddyutils.ParseBool(d, &n.ConfigOnly); err != nil {
				return err
			}
		case "check_duplicate":
			if err := caddyutils.ParseBool(d, &n.CheckDuplicate); err != nil {
				return err
			}
		case "driver_opts":
			n.Options = fnutils.DefaultIfEmptyMap(n.Options, make(map[string]string))
			if err := caddyutils.ParseKeyValuePairs(d, &n.Options, "="); err != nil {
				return err
			}
		case "label":
			n.Labels = fnutils.DefaultIfEmptyMap(n.Labels, make(map[string]string))
			var key, value string
			if err := caddyutils.ParseString(d, &key); err != nil {
				return err
			}
			if err := caddyutils.ParseString(d, &value); err != nil {
				return err
			}
			n.Labels[key] = value
		case "ipam_driver":
			n.IPAM = fnutils.DefaultIfNil(n.IPAM, new(IPAM))
			if err := caddyutils.ParseString(d, &n.IPAM.Driver); err != nil {
				return err
			}
		case "ipam":
			n.IPAM = fnutils.DefaultIfNil(n.IPAM, new(IPAM))
			n.IPAM.Config = fnutils.DefaultIfEmpty(n.IPAM.Config, []*IPAMConfig{})
			if n.IPAM.Driver == "" {
				n.IPAM.Driver = "default"
			}
			cfg := new(IPAMConfig)
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "subnet":
					if err := caddyutils.ParseString(d, &cfg.Subnet); err != nil {
						return err
					}
				case "gateway":
					if err := caddyutils.ParseString(d, &cfg.Gateway); err != nil {
						return err
					}
				case "ip_range":
					if err := caddyutils.ParseString(d, &cfg.IPRange); err != nil {
						return err
					}
				default:
					n.IPAM.Options = fnutils.DefaultIfEmptyMap(n.IPAM.Options, make(map[string]string))
					key := d.Val()
					var value string
					if err := caddyutils.ParseString(d, &value); err != nil {
						return err
					}
					n.IPAM.Options[key] = value
				}
			}
			n.IPAM.Config = append(n.IPAM.Config, cfg)
		default:
			return d.Errf("unknown option '%s'", d.Val())
		}
	}
	networks[name] = n
	return nil
}

func parseContainerOption(d *caddyfile.Dispenser, containers map[string]*Container) error {
	c := new(Container)
	name := ""
	if err := caddyutils.ParseString(d, &name); err != nil {
		return err
	}
	if name == "" {
		return d.Errf("container name is required")
	}
	_, ok := containers[name]
	if ok {
		return d.Errf("container '%s' already defined", name)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "image":
			if err := caddyutils.ParseString(d, &c.Image); err != nil {
				return err
			}
		case "hostname":
			if err := caddyutils.ParseString(d, &c.Hostname); err != nil {
				return err
			}
		case "domain":
			if err := caddyutils.ParseString(d, &c.Domainname); err != nil {
				return err
			}
		case "expose":
			c.ExposedPorts = fnutils.DefaultIfEmpty(c.ExposedPorts, []int{})
			if err := caddyutils.ParseIntArray(d, &c.ExposedPorts); err != nil {
				return err
			}
		case "volume", "volumes":
			c.Volumes = fnutils.DefaultIfEmpty(c.Volumes, []string{})
			if err := caddyutils.ParseStringArray(d, &c.Volumes, false); err != nil {
				return err
			}
		case "volume_from", "volumes_from":
			c.VolumesFrom = fnutils.DefaultIfEmpty(c.VolumesFrom, []string{})
			if err := caddyutils.ParseStringArray(d, &c.VolumesFrom, false); err != nil {
				return err
			}
		case "readonly_mount":
			c.Mounts = fnutils.DefaultIfEmpty(c.Mounts, []*Mount{})
			mount := &Mount{
				Type:     "bind",
				ReadOnly: true,
			}
			if err := caddyutils.ParseString(d, &mount.Source); err != nil {
				return err
			}
			if err := caddyutils.ExpectString(d, "to"); err != nil {
				return err
			}
			if err := caddyutils.ParseString(d, &mount.Target); err != nil {
				return err
			}
			c.Mounts = append(c.Mounts, mount)
		case "mount":
			c.Mounts = fnutils.DefaultIfEmpty(c.Mounts, []*Mount{})
			mount := &Mount{
				Type: "bind",
			}
			if d.CountRemainingArgs() > 0 {
				if err := caddyutils.ParseString(d, &mount.Source); err != nil {
					return err
				}
				if err := caddyutils.ExpectString(d, "to"); err != nil {
					return err
				}
				if err := caddyutils.ParseString(d, &mount.Target); err != nil {
					return err
				}
			} else {
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "type":
						if err := caddyutils.ParseString(d, &mount.Type); err != nil {
							return err
						}
					case "source":
						if err := caddyutils.ParseString(d, &mount.Source); err != nil {
							return err
						}
					case "target":
						if err := caddyutils.ParseString(d, &mount.Target); err != nil {
							return err
						}
					case "read_only":
						if err := caddyutils.ParseBool(d, &mount.ReadOnly); err != nil {
							return err
						}
					case "consistency":
						if err := caddyutils.ParseString(d, &mount.Consistency); err != nil {
							return err
						}
					default:
						return d.Errf("unknown option '%s'", d.Val())
					}
				}
			}
			c.Mounts = append(c.Mounts, mount)
		case "cmd", "command":
			if err := caddyutils.ParseStringArray(d, &c.Cmd, false); err != nil {
				return err
			}
		case "env", "environment":
			c.Env = fnutils.DefaultIfEmptyMap(c.Env, make(map[string]string))
			if d.CountRemainingArgs() > 0 {
				var key, value string
				if err := caddyutils.ParseString(d, &key); err != nil {
					return err
				}
				if err := caddyutils.ParseString(d, &value); err != nil {
					return err
				}
				c.Env[key] = value
			} else {
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					key := d.Val()
					var value string
					if err := caddyutils.ParseString(d, &value); err != nil {
						return err
					}
					c.Env[key] = value
				}
			}
		case "entrypoint":
			if err := caddyutils.ParseStringArray(d, &c.Entrypoint, false); err != nil {
				return err
			}
		case "healthcheck":
			c.Healthcheck = fnutils.DefaultIfNil(c.Healthcheck, new(HealthcheckConfig))
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "test":
					if err := caddyutils.ParseStringArray(d, &c.Healthcheck.Test, false); err != nil {
						return err
					}
				case "interval":
					if err := caddyutils.ParseDuration(d, &c.Healthcheck.Interval); err != nil {
						return err
					}
				case "timeout":
					if err := caddyutils.ParseDuration(d, &c.Healthcheck.Timeout); err != nil {
						return err
					}
				case "start_period":
					if err := caddyutils.ParseDuration(d, &c.Healthcheck.StartPeriod); err != nil {
						return err
					}
				case "retries":
					if err := caddyutils.ParseInt(d, &c.Healthcheck.Retries); err != nil {
						return err
					}
				default:
					return d.Errf("unknown option '%s'", d.Val())
				}
			}
		case "label":
			c.Labels = fnutils.DefaultIfEmptyMap(c.Labels, make(map[string]string))
			var key, value string
			if err := caddyutils.ParseString(d, &key); err != nil {
				return err
			}
			if err := caddyutils.ParseString(d, &value); err != nil {
				return err
			}
			c.Labels[key] = value
		case "user":
			if err := caddyutils.ParseString(d, &c.User); err != nil {
				return err
			}
		case "restart":
			policy := RestartPolicy{}
			if d.CountRemainingArgs() > 0 {
				if err := caddyutils.ParseString(d, &policy.Name); err != nil {
					return err
				}
				if d.CountRemainingArgs() > 0 {
					if err := caddyutils.ParseInt(d, &policy.Max); err != nil {
						return err
					}
				}
			} else {
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "name":
						if err := caddyutils.ParseString(d, &c.RestartPolicy.Name); err != nil {
							return err
						}
					case "max":
						if err := caddyutils.ParseInt(d, &c.RestartPolicy.Max); err != nil {
							return err
						}
					default:
						return d.Errf("unknown option '%s'", d.Val())
					}
				}
			}
			c.RestartPolicy = &policy
		case "stop_signal":
			if err := caddyutils.ParseString(d, &c.StopSignal); err != nil {
				return err
			}
		case "stop_timeout":
			if err := caddyutils.ParseSecondsDuration(d, c.StopTimeout); err != nil {
				return err
			}
		case "network":
			c.Networks = fnutils.DefaultIfEmpty(c.Networks, []*NetworkConfig{})
			network := NetworkConfig{}
			if d.CountRemainingArgs() > 0 {
				if err := caddyutils.ParseString(d, &network.NetworkID); err != nil {
					return err
				}
			} else {
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "name":
						if err := caddyutils.ParseString(d, &network.NetworkID); err != nil {
							return err
						}
					case "endpoint":
						if err := caddyutils.ParseString(d, &network.EndpointID); err != nil {
							return err
						}
					case "gateway":
						if err := caddyutils.ParseString(d, &network.Gateway); err != nil {
							return err
						}
					case "ip_address":
						if err := caddyutils.ParseString(d, &network.IPAddress); err != nil {
							return err
						}
					case "ip_prefix_length":
						if err := caddyutils.ParseInt(d, &network.IPPrefixLen); err != nil {
							return err
						}
					case "ipv6_gateway":
						if err := caddyutils.ParseString(d, &network.IPv6Gateway); err != nil {
							return err
						}
					case "ipv6_address":
						if err := caddyutils.ParseString(d, &network.GlobalIPv6Address); err != nil {
							return err
						}
					case "ipv6_prefix_length":
						if err := caddyutils.ParseInt(d, &network.GlobalIPv6PrefixLen); err != nil {
							return err
						}
					case "mac_address":
						if err := caddyutils.ParseString(d, &network.MacAddress); err != nil {
							return err
						}
					case "link":
						network.Links = fnutils.DefaultIfEmpty(network.Links, []string{})
						if err := caddyutils.ParseStringArray(d, &network.Links, false); err != nil {
							return err
						}
					case "alias":
						network.Aliases = fnutils.DefaultIfEmpty(network.Aliases, []string{})
						if err := caddyutils.ParseStringArray(d, &network.Aliases, false); err != nil {
							return err
						}
					case "driver_opts":
						network.DriverOpts = fnutils.DefaultIfEmptyMap(network.DriverOpts, make(map[string]string))
						if err := caddyutils.ParseKeyValuePairs(d, &network.DriverOpts, "="); err != nil {
							return err
						}
					default:
						return d.Errf("unknown option '%s'", d.Val())
					}
				}
			}
			c.Networks = append(c.Networks, &network)
		default:
			return d.Errf("unknown option '%s'", d.Val())
		}

	}
	containers[name] = c
	return nil
}
