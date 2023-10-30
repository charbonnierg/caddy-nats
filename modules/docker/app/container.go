package app

import (
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/pkg/datatypes"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

// EndpointIPAMConfig represents IPAM configurations for the endpoint
type EndpointIPAMConfig struct {
	IPv4Address  string   `json:"ipv4_address,omitempty"`
	IPv6Address  string   `json:"ipv6_address,omitempty"`
	LinkLocalIPs []string `json:"link_local_ips,omitempty"`
}

type Mount struct {
	Type           string                `json:",omitempty"`
	Source         string                `json:"source,omitempty"`
	Target         string                `json:"target,omitempty"`
	ReadOnly       bool                  `json:"readonly,omitempty"`
	Consistency    string                `json:"consistency,omitempty"`
	BindOptions    *mount.BindOptions    `json:"bind_opts,omitempty"`
	VolumeOptions  *mount.VolumeOptions  `json:"volume_opts,omitempty"`
	TmpfsOptions   *mount.TmpfsOptions   `json:"tmpfs_opts,omitempty"`
	ClusterOptions *mount.ClusterOptions `json:"cluster_opts,omitempty"`
}

type PortBinding struct {
	HostAddress   string `json:"host_address,omitempty"`
	HostPort      int    `json:"host,omitempty"`
	ContainerPort int    `json:"container,omitempty"`
}

// HealthcheckConfig holds configuration settings for the HEALTHCHECK feature.
type HealthcheckConfig struct {
	Test        []string      `json:"test,omitempty"`
	Interval    time.Duration `json:"interval,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
	StartPeriod time.Duration `json:"start_period,omitempty"`
	Retries     int           `json:"retries,omitempty"`
}

type LogConfig struct {
	Type   string            `json:"type,omitempty"`
	Config map[string]string `json:"config,omitempty"`
}

type RestartPolicy struct {
	Name string `json:"policy,omitempty"`
	Max  int    `json:"max,omitempty"`
}

type Ulimit struct {
	Name string `json:"name,omitempty"`
	Hard int64  `json:"hard,omitempty"`
	Soft int64  `json:"soft,omitempty"`
}

type DeviceMapping struct {
	PathOnHost        string `json:"path_on_host,omitempty"`
	PathInContainer   string `json:"path_in_container,omitempty"`
	CgroupPermissions string `json:"cgroup_permissions,omitempty"`
}

type DeviceRequest struct {
	Driver       string            `json:"driver,omitempty"`       // Name of device driver
	Count        int               `json:"count,omitempty"`        // Number of devices to request (-1 = All)
	DeviceIDs    []string          `json:"device_ids,omitempty"`   // List of device IDs as recognizable by the device driver
	Capabilities [][]string        `json:"capabilities,omitempty"` // An OR list of AND lists of device capabilities (e.g. "gpu")
	Options      map[string]string `json:"options,omitempty"`      // Options to pass onto the device driver
}

type WeightDevice struct {
	Path   string `json:"path,omitempty"`
	Weight uint16 `json:"weight,omitempty"`
}

type ThrottleDevice struct {
	Path string `json:"path,omitempty"`
	Rate uint64 `json:"rate,omitempty"`
}

// Resources contains container's resources (cgroups config, ulimits...)
type Resources struct {
	// Applicable to all platforms
	CPUShares int64 `json:"cpu_shares,omitempty"` // CPU shares (relative weight vs. other containers)
	Memory    int64 `json:"memory,omitempty"`     // Memory limit (in bytes)
	NanoCPUs  int64 `json:"nano_cpus,omitempty"`  // CPU quota in units of 10<sup>-9</sup> CPUs.

	// Applicable to UNIX platforms
	CgroupParent         string            `json:"cgroup_parent,omitempty"` // Parent cgroup.
	BlkioWeight          uint16            `json:"blkio_weight,omitempty"`  // Block IO weight (relative weight vs. other containers)
	BlkioWeightDevice    []*WeightDevice   `json:"blkio_weight_device,omitempty"`
	BlkioDeviceReadBps   []*ThrottleDevice `json:"blkio_device_read_bps,omitempty"`
	BlkioDeviceWriteBps  []*ThrottleDevice `json:"blkio_device_write_bps,omitempty"`
	BlkioDeviceReadIOps  []*ThrottleDevice `json:"blkio_device_read_iops,omitempty"`
	BlkioDeviceWriteIOps []*ThrottleDevice `json:"blkio_device_write_iops,omitempty"`
	CPUPeriod            int64             `json:"cpu_period,omitempty"`           // CPU CFS (Completely Fair Scheduler) period
	CPUQuota             int64             `json:"cpu_quota,omitempty"`            // CPU CFS (Completely Fair Scheduler) quota
	CPURealtimePeriod    int64             `json:"cpu_realtime_period,omitempty"`  // CPU real-time period
	CPURealtimeRuntime   int64             `json:"cpu_realtime_runtime,omitempty"` // CPU real-time runtime
	CpusetCpus           string            `json:"cpu_set_cpus,omitempty"`         // CpusetCpus 0-2, 0,1
	CpusetMems           string            `json:"cpu_set_mems,omitempty"`         // CpusetMems 0-2, 0,1
	Devices              []DeviceMapping   `json:"devices,omitempty"`              // List of devices to map inside the container
	DeviceCgroupRules    []string          `json:"device_cgroup_rules,omitempty"`  // List of rule to be added to the device cgroup
	DeviceRequests       []DeviceRequest   `json:"device_requests,omitempty"`      // List of device requests for device drivers

	// KernelMemory specifies the kernel memory limit (in bytes) for the container.
	// Deprecated: kernel 5.4 deprecated kmem.limit_in_bytes.
	KernelMemory      int64     `json:"kernel_memory,omitempty"`
	KernelMemoryTCP   int64     `json:"kernel_memory_tcp,omitempty"`  // Hard limit for kernel TCP buffer memory (in bytes)
	MemoryReservation int64     `json:"memory_reservation,omitempty"` // Memory soft limit (in bytes)
	MemorySwap        int64     `json:"memory_swap,omitempty"`        // Total memory usage (memory + swap); set `-1` to enable unlimited swap
	MemorySwappiness  *int64    `json:"memory_swappiness,omitempty"`  // Tuning container memory swappiness behaviour
	OomKillDisable    *bool     `json:"oom_kill_disable,omitempty"`   // Whether to disable OOM Killer or not
	PidsLimit         *int64    `json:"pids_limit,omitempty"`         // Setting PIDs limit for a container; Set `0` or `-1` for unlimited, or `null` to not change.
	Ulimits           []*Ulimit `json:"ulimits,omitempty"`            // List of ulimits to be set in the container

	// Applicable to Windows
	CPUCount           int64  `json:"win_cpu_count,omitempty"`        // CPU count
	CPUPercent         int64  `json:"win_cpu_percent,omitempty"`      // CPU percent
	IOMaximumIOps      uint64 `json:"win_maximum_iops,omitempty"`     // Maximum IOps for the container system drive
	IOMaximumBandwidth uint64 `json:"win_maximum_bandwith,omitempty"` // Maximum IO in bytes per second for the container system drive
}

type HostConfig struct {
	// Applicable to all platforms
	Binds           []string          `json:"binds,omitempty"`             // List of volume bindings for this container
	ContainerIDFile string            `json:"container_id_file,omitempty"` // File (path) where the containerId is written
	LogConfig       *LogConfig        `json:"log_config,omitempty"`        // Configuration of the logs for this container
	NetworkMode     string            `json:"network,omitempty"`           // Network mode to use for the container
	PortBindings    []*PortBinding    `json:"ports,omitempty"`             // Port mapping between the exposed port (container) and the host
	RestartPolicy   *RestartPolicy    `json:"restart,omitempty"`           // Restart policy to be used for the container
	AutoRemove      bool              `json:"auto_remove,omitempty"`       // Automatically remove container when it exits
	VolumeDriver    string            `json:"volume_driver,omitempty"`     // Name of the volume driver used to mount volumes
	VolumesFrom     []string          `json:"volumes_from,omitempty"`      // List of volumes to take from other container
	Annotations     map[string]string `json:"annotations,omitempty"`       // Arbitrary non-identifying metadata attached to container and provided to the runtime

	// Applicable to UNIX platforms
	CapAdd          []string          `json:"cap_add,omitempty"`           // List of kernel capabilities to add to the container
	CapDrop         []string          `json:"cap_drop,omitempty"`          // List of kernel capabilities to remove from the container
	CgroupnsMode    string            `json:"cgroup_ns_mode,omitempty"`    // Cgroup namespace mode to use for the container
	DNS             []string          `json:"dns,omitempty"`               // List of DNS server to lookup
	DNSOptions      []string          `json:"dns_options,omitempty"`       // List of DNSOption to look for
	DNSSearch       []string          `json:"dns_search,omitempty"`        // List of DNSSearch to look for
	ExtraHosts      []string          `json:"extra_hosts,omitempty"`       // List of extra hosts
	GroupAdd        []string          `json:"group_add,omitempty"`         // List of additional groups that the container process will run as
	IpcMode         string            `json:"ipc_mode,omitempty"`          // IPC namespace to use for the container
	Cgroup          string            `json:"cgroup,omitempty"`            // Cgroup to use for the container
	Links           []string          `json:"links,omitempty"`             // List of links (in the name:alias form)
	OomScoreAdj     int               `json:"oom_score_adj,omitempty"`     // Container preference for OOM-killing
	PidMode         string            `json:"pid_mode,omitempty"`          // PID namespace to use for the container
	Privileged      bool              `json:"privileged,omitempty"`        // Is the container in privileged mode
	PublishAllPorts bool              `json:"publish_all_ports,omitempty"` // Should docker publish all exposed port for the container
	ReadonlyRootfs  bool              `json:"readonly_root_fs,omitempty"`  // Is the container root filesystem read-only
	SecurityOpt     []string          `json:"security"`                    // List of string values to customize labels for MLS systems, such as SELinux.
	StorageOpt      map[string]string `json:"storage,omitempty"`           // Storage driver options per container.
	Tmpfs           map[string]string `json:"tmpfs,omitempty"`             // List of tmpfs (mounts) used for the container
	UTSMode         string            `json:"uts_mode,omitempty"`          // UTS namespace to use for the container
	UsernsMode      string            `json:"user_ns_mode,omitempty"`      // The user namespace to use for the container
	ShmSize         int64             `json:"shm_size,omitempty"`          // Total shm memory usage
	Sysctls         map[string]string `json:"sysctls,omitempty"`           // List of Namespaced sysctls used for the container
	Runtime         string            `json:"runtime,omitempty"`           // Runtime to use with this container

	// Applicable to Windows
	Isolation string `json:"isolation,omitempty"` // Isolation technology of the container (e.g. default, hyperv)

	// Applicable to UNIX and Windows platforms
	Mounts        []*Mount `json:"mounts,omitempty"`         // Mounts specs used by the container
	MaskedPaths   []string `json:"masked_paths,omitempty"`   // MaskedPaths is the list of paths to be masked inside the container (this overrides the default set of paths)
	ReadonlyPaths []string `json:"readonly_paths,omitempty"` // ReadonlyPaths is the list of paths to be set as read-only inside the container (this overrides the default set of paths)
	Init          *bool    `json:"init,omitempty"`           // Run a custom init inside the container, if null, use the daemon's configured settings
}

type NetworkConfig struct {
	// Configurations
	Links   []string `json:"links,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	// Operational data
	NetworkID           string              `json:"id,omitempty"`
	EndpointID          string              `json:"endpoint_id,omitempty"`
	Gateway             string              `json:"gateway,omitempty"`
	IPAMConfig          *EndpointIPAMConfig `json:"ipam_config,omitempty"`
	IPAddress           string              `json:"ip_address,omitempty"`
	IPPrefixLen         int                 `json:"ip_prefix_len,omitempty"`
	IPv6Gateway         string              `json:"ipv6_gateway,omitempty"`
	GlobalIPv6Address   string              `json:"global_ipv6_address,omitempty"`
	GlobalIPv6PrefixLen int                 `json:"global_ipv6_prefix_len,omitempty"`
	MacAddress          string              `json:"mac_address,omitempty"`
	DriverOpts          map[string]string   `json:"driver_opts,omitempty"`
}

type Container struct {
	HostConfig
	Resources
	Image           string             `json:"image"`
	Hostname        string             `json:"hostname,omitempty"`
	Domainname      string             `json:"domain,omitempty"`
	User            string             `json:"user,omitempty"`
	ExposedPorts    []int              `json:"ports,omitempty"`
	Env             map[string]string  `json:"env,omitempty"`
	Cmd             []string           `json:"cmd,omitempty"`
	Healthcheck     *HealthcheckConfig `json:"healthcheck,omitempty"`
	Volumes         []string           `json:"volumes,omitempty"`
	WorkingDir      string             `json:"working_dir,omitempty"`
	Entrypoint      []string           `json:"entrypoint,omitempty"`
	NetworkDisabled bool               `json:"network_disabled,omitempty"`
	MacAddress      string             `json:"mac_address,omitempty"`
	Labels          map[string]string  `json:"labels,omitempty"`
	StopSignal      string             `json:"stop_signal,omitempty"`
	StopTimeout     *int               `json:"stop_timeout,omitempty"`
	Networks        []*NetworkConfig   `json:"networks,omitempty"`
}

func (c *Container) logConfig() container.LogConfig {
	cfg := container.LogConfig{}
	if c.LogConfig == nil {
		return cfg
	}
	cfg.Type = c.LogConfig.Type
	cfg.Config = c.LogConfig.Config
	return cfg
}

func (c *Container) portBindings() nat.PortMap {
	portmap := nat.PortMap{}
	if c.PortBindings == nil {
		return portmap
	}
	for _, p := range c.PortBindings {
		portmap[nat.Port(fmt.Sprintf("%d/tcp", p.ContainerPort))] = []nat.PortBinding{
			{
				HostIP:   p.HostAddress,
				HostPort: fmt.Sprintf("%d", p.HostPort),
			},
		}
	}
	return portmap
}

func (c *Container) exposedPorts() nat.PortSet {
	ports := nat.PortSet{}
	for _, p := range c.ExposedPorts {
		ports[nat.Port(fmt.Sprintf("%d/tcp", p))] = struct{}{}
	}
	return ports
}

func (c *Container) restartPolicy() container.RestartPolicy {
	policy := container.RestartPolicy{}
	if c.RestartPolicy == nil {
		return policy
	}
	policy.Name = c.RestartPolicy.Name
	policy.MaximumRetryCount = c.RestartPolicy.Max
	return policy
}

func (c *Container) healthcheck() *container.HealthConfig {
	if c.Healthcheck == nil {
		return nil
	}
	return &container.HealthConfig{
		Test:        c.Healthcheck.Test,
		Interval:    c.Healthcheck.Interval,
		Timeout:     c.Healthcheck.Timeout,
		StartPeriod: c.Healthcheck.StartPeriod,
		Retries:     c.Healthcheck.Retries,
	}
}

func (c *Container) mounts() []mount.Mount {
	allMounts := make([]mount.Mount, len(c.Mounts))
	for idx, m := range c.Mounts {
		allMounts[idx] = mount.Mount{
			Type:           mount.Type(m.Type),
			Source:         m.Source,
			Target:         m.Target,
			ReadOnly:       m.ReadOnly,
			Consistency:    mount.Consistency(m.Consistency),
			BindOptions:    m.BindOptions,
			VolumeOptions:  m.VolumeOptions,
			TmpfsOptions:   m.TmpfsOptions,
			ClusterOptions: m.ClusterOptions,
		}
	}
	return allMounts
}

func (c *Container) volumes() map[string]struct{} {
	volumes := datatypes.StringSet{}
	for _, v := range c.Volumes {
		volumes.Add(v)
	}
	return volumes
}

func (c *Container) environ(repl *caddy.Replacer, cfg *container.Config) error {
	environ := datatypes.Environ{}
	for k, v := range c.Env {
		v, err := repl.ReplaceOrErr(v, true, true)
		if err != nil {
			return err
		}
		environ.Set(k, v)
	}
	cfg.Env = fnutils.DefaultIfEmpty(cfg.Env, []string{})
	cfg.Env = append(cfg.Env, environ.Entries()...)
	return nil
}

func (c *Container) blkioWeightDevices() []*blkiodev.WeightDevice {
	devices := make([]*blkiodev.WeightDevice, len(c.BlkioWeightDevice))
	for idx, d := range c.BlkioWeightDevice {
		devices[idx] = &blkiodev.WeightDevice{
			Path:   d.Path,
			Weight: d.Weight,
		}
	}
	return devices
}

func (c *Container) blkioDeviceReadBps() []*blkiodev.ThrottleDevice {
	devices := make([]*blkiodev.ThrottleDevice, len(c.BlkioDeviceReadBps))
	for idx, d := range c.BlkioDeviceReadBps {
		devices[idx] = &blkiodev.ThrottleDevice{
			Path: d.Path,
			Rate: d.Rate,
		}
	}
	return devices
}

func (c *Container) blkioDeviceWriteBps() []*blkiodev.ThrottleDevice {
	devices := make([]*blkiodev.ThrottleDevice, len(c.BlkioDeviceWriteBps))
	for idx, d := range c.BlkioDeviceWriteBps {
		devices[idx] = &blkiodev.ThrottleDevice{
			Path: d.Path,
			Rate: d.Rate,
		}
	}
	return devices
}

func (c *Container) blkioDeviceReadIOps() []*blkiodev.ThrottleDevice {
	devices := make([]*blkiodev.ThrottleDevice, len(c.BlkioDeviceReadIOps))
	for idx, d := range c.BlkioDeviceReadIOps {
		devices[idx] = &blkiodev.ThrottleDevice{
			Path: d.Path,
			Rate: d.Rate,
		}
	}
	return devices
}

func (c *Container) blkioDeviceWriteIOps() []*blkiodev.ThrottleDevice {
	devices := make([]*blkiodev.ThrottleDevice, len(c.BlkioDeviceWriteIOps))
	for idx, d := range c.BlkioDeviceWriteIOps {
		devices[idx] = &blkiodev.ThrottleDevice{
			Path: d.Path,
			Rate: d.Rate,
		}
	}
	return devices
}

func (c *Container) devices() []container.DeviceMapping {
	devices := make([]container.DeviceMapping, len(c.Devices))
	for idx, d := range c.Devices {
		devices[idx] = container.DeviceMapping{
			PathOnHost:        d.PathOnHost,
			PathInContainer:   d.PathInContainer,
			CgroupPermissions: d.CgroupPermissions,
		}
	}
	return devices
}

func (c *Container) deviceRequests() []container.DeviceRequest {
	requests := make([]container.DeviceRequest, len(c.DeviceRequests))
	for idx, r := range c.DeviceRequests {
		requests[idx] = container.DeviceRequest{
			Driver:       r.Driver,
			Count:        r.Count,
			DeviceIDs:    r.DeviceIDs,
			Capabilities: r.Capabilities,
			Options:      r.Options,
		}
	}
	return requests
}

func (c *Container) ulimits() []*units.Ulimit {
	ulimits := make([]*units.Ulimit, len(c.Ulimits))
	for idx, u := range c.Ulimits {
		ulimits[idx] = &units.Ulimit{
			Name: u.Name,
			Hard: u.Hard,
			Soft: u.Soft,
		}
	}
	return ulimits
}

func (c *Container) ContainerConfig(repl *caddy.Replacer) (*container.Config, error) {
	cfg := container.Config{}
	if err := replaceOrErr(repl, c.Image, &cfg.Image); err != nil {
		return nil, err
	}
	if err := replaceOrErr(repl, c.Hostname, &cfg.Hostname); err != nil {
		return nil, err
	}
	if err := replaceOrErr(repl, c.Domainname, &cfg.Domainname); err != nil {
		return nil, err
	}
	if err := replaceOrErr(repl, c.User, &cfg.User); err != nil {
		return nil, err
	}
	cfg.ExposedPorts = c.exposedPorts()
	if err := c.environ(repl, &cfg); err != nil {
		return nil, err
	}
	cfg.Cmd = c.Cmd
	cfg.Healthcheck = c.healthcheck()
	cfg.Volumes = c.volumes()
	if err := replaceOrErr(repl, c.WorkingDir, &cfg.WorkingDir); err != nil {
		return nil, err
	}
	cfg.Entrypoint = c.Entrypoint
	cfg.NetworkDisabled = c.NetworkDisabled
	cfg.MacAddress = c.MacAddress
	cfg.Labels = c.Labels
	cfg.StopSignal = c.StopSignal
	cfg.StopTimeout = c.StopTimeout
	return &cfg, nil
}

func (c *Container) ContainerHostConfig(repl *caddy.Replacer) (*container.HostConfig, error) {
	return &container.HostConfig{
		Binds:           c.Binds,
		ContainerIDFile: c.ContainerIDFile,
		LogConfig:       c.logConfig(),
		NetworkMode:     container.NetworkMode(c.NetworkMode),
		PortBindings:    c.portBindings(),
		RestartPolicy:   c.restartPolicy(),
		AutoRemove:      c.AutoRemove,
		VolumeDriver:    c.VolumeDriver,
		VolumesFrom:     c.VolumesFrom,
		Annotations:     c.Annotations,
		CapAdd:          c.CapAdd,
		CapDrop:         c.CapDrop,
		CgroupnsMode:    container.CgroupnsMode(c.CgroupnsMode),
		DNS:             c.DNS,
		DNSOptions:      c.DNSOptions,
		DNSSearch:       c.DNSSearch,
		ExtraHosts:      c.ExtraHosts,
		GroupAdd:        c.GroupAdd,
		IpcMode:         container.IpcMode(c.IpcMode),
		Cgroup:          container.CgroupSpec(c.Cgroup),
		Links:           c.Links,
		OomScoreAdj:     c.OomScoreAdj,
		PidMode:         container.PidMode(c.PidMode),
		Privileged:      c.Privileged,
		PublishAllPorts: c.PublishAllPorts,
		ReadonlyRootfs:  c.ReadonlyRootfs,
		SecurityOpt:     c.SecurityOpt,
		StorageOpt:      c.StorageOpt,
		Tmpfs:           c.Tmpfs,
		UTSMode:         container.UTSMode(c.UTSMode),
		UsernsMode:      container.UsernsMode(c.UsernsMode),
		ShmSize:         c.ShmSize,
		Sysctls:         c.Sysctls,
		Runtime:         c.Runtime,
		Isolation:       container.Isolation(c.Isolation),
		Mounts:          c.mounts(),
		MaskedPaths:     c.MaskedPaths,
		ReadonlyPaths:   c.ReadonlyPaths,
		Init:            c.Init,
		Resources: container.Resources{
			CPUShares:            c.CPUShares,
			Memory:               c.Memory,
			NanoCPUs:             c.NanoCPUs,
			CgroupParent:         c.CgroupParent,
			BlkioWeight:          c.BlkioWeight,
			BlkioWeightDevice:    c.blkioWeightDevices(),
			BlkioDeviceReadBps:   c.blkioDeviceReadBps(),
			BlkioDeviceWriteBps:  c.blkioDeviceWriteBps(),
			BlkioDeviceReadIOps:  c.blkioDeviceReadIOps(),
			BlkioDeviceWriteIOps: c.blkioDeviceWriteIOps(),
			CPUPeriod:            c.CPUPeriod,
			CPUQuota:             c.CPUQuota,
			CPURealtimePeriod:    c.CPURealtimePeriod,
			CPURealtimeRuntime:   c.CPURealtimeRuntime,
			CpusetCpus:           c.CpusetCpus,
			CpusetMems:           c.CpusetMems,
			Devices:              c.devices(),
			DeviceCgroupRules:    c.DeviceCgroupRules,
			DeviceRequests:       c.deviceRequests(),
			KernelMemory:         c.KernelMemory,
			KernelMemoryTCP:      c.KernelMemoryTCP,
			MemoryReservation:    c.MemoryReservation,
			MemorySwap:           c.MemorySwap,
			MemorySwappiness:     c.MemorySwappiness,
			OomKillDisable:       c.OomKillDisable,
			PidsLimit:            c.PidsLimit,
			Ulimits:              c.ulimits(),
			CPUCount:             c.CPUCount,
			CPUPercent:           c.CPUPercent,
			IOMaximumIOps:        c.IOMaximumIOps,
			IOMaximumBandwidth:   c.IOMaximumBandwidth,
		},
	}, nil
}

func (c *Container) ContainerNetworkConfig(repl *caddy.Replacer) (*network.NetworkingConfig, error) {
	cfg := network.NetworkingConfig{}
	endpoints := make(map[string]*network.EndpointSettings)
	for _, n := range c.Networks {
		ipam := &network.EndpointIPAMConfig{}
		if n.IPAMConfig != nil {
			ipam.IPv4Address = n.IPAMConfig.IPv4Address
			ipam.IPv6Address = n.IPAMConfig.IPv6Address
			ipam.LinkLocalIPs = n.IPAMConfig.LinkLocalIPs
		} else {
			ipam = nil
		}
		endpoints[n.NetworkID] = &network.EndpointSettings{
			NetworkID:           n.NetworkID,
			EndpointID:          n.EndpointID,
			Gateway:             n.Gateway,
			IPAMConfig:          ipam,
			IPAddress:           n.IPAddress,
			IPPrefixLen:         n.IPPrefixLen,
			IPv6Gateway:         n.IPv6Gateway,
			GlobalIPv6Address:   n.GlobalIPv6Address,
			GlobalIPv6PrefixLen: n.GlobalIPv6PrefixLen,
			MacAddress:          n.MacAddress,
			DriverOpts:          n.DriverOpts,
		}
	}
	cfg.EndpointsConfig = endpoints
	return &cfg, nil
}

var (
	_ docker.Container = (*Container)(nil)
)

func replaceOrErr(repl *caddy.Replacer, in string, dest *string) error {
	out, err := repl.ReplaceOrErr(in, true, true)
	if err != nil {
		return err
	}
	*dest = out
	return nil
}
