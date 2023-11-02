package hostmetrics

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (r *HostMetricsReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "hostmetrics"); err != nil {
		return err
	}
	r.Scrapers = fnutils.DefaultIfNil(r.Scrapers, &Scrapers{})
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "collection_interval":
			if err := caddyutils.ParseDuration(d, &r.CollectionInterval); err != nil {
				return err
			}
		case "initial_delay":
			if err := caddyutils.ParseDuration(d, &r.InitialDelay); err != nil {
				return err
			}
		case "root_path":
			if err := caddyutils.ParseString(d, &r.RootPath); err != nil {
				return err
			}
		case "scrap":
			for d.NextArg() {
				switch d.Val() {
				case "cpu":
					r.Scrapers.Cpu = fnutils.DefaultIfNil(r.Scrapers.Cpu, &CpuScraper{})
				case "disk":
					r.Scrapers.Disk = fnutils.DefaultIfNil(r.Scrapers.Disk, &DiskScraper{})
				case "load":
					r.Scrapers.Load = fnutils.DefaultIfNil(r.Scrapers.Load, &LoadScraper{})
				case "filesystem":
					r.Scrapers.Filesystem = fnutils.DefaultIfNil(r.Scrapers.Filesystem, &FilesystemScraper{})
				case "memory":
					r.Scrapers.Memory = fnutils.DefaultIfNil(r.Scrapers.Memory, &MemoryScrapper{})
				case "network":
					r.Scrapers.Network = fnutils.DefaultIfNil(r.Scrapers.Network, &NetworkScrapper{})
				case "paging":
					r.Scrapers.Paging = fnutils.DefaultIfNil(r.Scrapers.Paging, &PagingScraper{})
				case "processes":
					r.Scrapers.Processes = fnutils.DefaultIfNil(r.Scrapers.Processes, &ProcessesScraper{})
				case "process":
					r.Scrapers.Process = fnutils.DefaultIfNil(r.Scrapers.Process, &ProcessScapper{})
				default:
					return d.Errf("unrecognized scraper %s", d.Val())
				}
			}
		case "cpu":
			r.Scrapers.Cpu = fnutils.DefaultIfNil(r.Scrapers.Cpu, &CpuScraper{})
			if err := r.unmarshalCpuScrapper(d, r.Scrapers.Cpu); err != nil {
				return err
			}
		case "disk":
			r.Scrapers.Disk = fnutils.DefaultIfNil(r.Scrapers.Disk, &DiskScraper{})
			if err := r.unmarshalDiskScrapper(d, r.Scrapers.Disk); err != nil {
				return err
			}
		case "load":
			r.Scrapers.Load = fnutils.DefaultIfNil(r.Scrapers.Load, &LoadScraper{})
			if err := r.unmarshalLoadScrapper(d, r.Scrapers.Load); err != nil {
				return err
			}
		case "filesystem":
			r.Scrapers.Filesystem = fnutils.DefaultIfNil(r.Scrapers.Filesystem, &FilesystemScraper{})
			if err := r.unmarshalFilesystemScrapper(d, r.Scrapers.Filesystem); err != nil {
				return err
			}
		case "memory":
			r.Scrapers.Memory = fnutils.DefaultIfNil(r.Scrapers.Memory, &MemoryScrapper{})
			if err := r.unmarshalMemoryScrapper(d, r.Scrapers.Memory); err != nil {
				return err
			}
		case "network":
			r.Scrapers.Network = fnutils.DefaultIfNil(r.Scrapers.Network, &NetworkScrapper{})
			if err := r.unmarshalNetworkScrapper(d, r.Scrapers.Network); err != nil {
				return err
			}
		case "paging":
			r.Scrapers.Paging = fnutils.DefaultIfNil(r.Scrapers.Paging, &PagingScraper{})
			if err := r.unmarshalPagingScrapper(d, r.Scrapers.Paging); err != nil {
				return err
			}
		case "processes":
			r.Scrapers.Processes = fnutils.DefaultIfNil(r.Scrapers.Processes, &ProcessesScraper{})
			if err := r.unmarshalProcessesScrapper(d, r.Scrapers.Processes); err != nil {
				return err
			}
		case "process":
			r.Scrapers.Process = fnutils.DefaultIfNil(r.Scrapers.Process, &ProcessScapper{})
			if err := r.unmarshalProcessScrapper(d, r.Scrapers.Process); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalCpuScrapper(d *caddyfile.Dispenser, s *CpuScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalDiskScrapper(d *caddyfile.Dispenser, s *DiskScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include_device", "include_devices":
			s.IncludeDevices = &DeviceFilter{}
			if err := caddyutils.ParseString(d, &s.IncludeDevices.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.IncludeDevices.DeviceNames, false); err != nil {
				return err
			}
		case "exclude_device", "exclude_devices":
			s.ExcludeDevices = &DeviceFilter{}
			if err := caddyutils.ParseString(d, &s.ExcludeDevices.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.ExcludeDevices.DeviceNames, false); err != nil {
				return err
			}
		case "include_fs_type", "include_fs_types":
			s.IncludeFSTypes = &FSTypeFilter{}
			if err := caddyutils.ParseString(d, &s.IncludeFSTypes.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.IncludeFSTypes.FSTypes, false); err != nil {
				return err
			}
		case "exclude_fs_type", "exclude_fs_types":
			s.ExcludeFSTypes = &FSTypeFilter{}
			if err := caddyutils.ParseString(d, &s.ExcludeFSTypes.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.ExcludeFSTypes.FSTypes, false); err != nil {
				return err
			}
		case "include_mount_point", "include_mount_points":
			s.IncludeMountPoints = &MountPointFilter{}
			if err := caddyutils.ParseString(d, &s.IncludeMountPoints.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.IncludeMountPoints.MountPoints, false); err != nil {
				return err
			}
		case "exclude_mount_point", "exclude_mount_points":
			s.ExcludeMountPoints = &MountPointFilter{}
			if err := caddyutils.ParseString(d, &s.ExcludeMountPoints.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.ExcludeMountPoints.MountPoints, false); err != nil {
				return err
			}
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalLoadScrapper(d *caddyfile.Dispenser, s *LoadScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "cpu_average":
			if err := caddyutils.ParseBool(d, &s.CpuAverage); err != nil {
				return err
			}
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalFilesystemScrapper(d *caddyfile.Dispenser, s *FilesystemScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalMemoryScrapper(d *caddyfile.Dispenser, s *MemoryScrapper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalNetworkScrapper(d *caddyfile.Dispenser, s *NetworkScrapper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include":
			s.IncludeInterfaces = &NetworkInterfaceFilter{}
			if err := caddyutils.ParseString(d, &s.IncludeInterfaces.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.IncludeInterfaces.Interfaces, false); err != nil {
				return err
			}
		case "exclude":
			s.ExcludeInterfaces = &NetworkInterfaceFilter{}
			if err := caddyutils.ParseString(d, &s.ExcludeInterfaces.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.ExcludeInterfaces.Interfaces, false); err != nil {
				return err
			}
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalPagingScrapper(d *caddyfile.Dispenser, s *PagingScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalProcessesScrapper(d *caddyfile.Dispenser, s *ProcessesScraper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *HostMetricsReceiver) unmarshalProcessScrapper(d *caddyfile.Dispenser, s *ProcessScapper) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include":
			s.IncludeProcesses = &ProcessNameFilter{}
			if err := caddyutils.ParseString(d, &s.IncludeProcesses.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.IncludeProcesses.Names, false); err != nil {
				return err
			}
		case "exclude":
			s.ExcludeProcesses = &ProcessNameFilter{}
			if err := caddyutils.ParseString(d, &s.ExcludeProcesses.MatchType); err != nil {
				return err
			}
			if err := caddyutils.ParseStringArray(d, &s.ExcludeProcesses.Names, false); err != nil {
				return err
			}
		case "mute_process_name_error":
			if err := caddyutils.ParseBool(d, &s.MuteProcessNameError); err != nil {
				return err
			}
		case "mute_process_exe_error":
			if err := caddyutils.ParseBool(d, &s.MuteProcessExeError); err != nil {
				return err
			}
		case "mute_process_io_error":
			if err := caddyutils.ParseBool(d, &s.MuteProcessIOErrror); err != nil {
				return err
			}
		case "mute_process_user_error":
			if err := caddyutils.ParseBool(d, &s.MuteProcessUserError); err != nil {
				return err
			}
		case "scrap_process_delay":
			if err := caddyutils.ParseDuration(d, &s.ScrapeProcessDelay); err != nil {
				return err
			}
		default:
			if err := parseMetricsEnabled(d, &s.Metrics); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseMetricsEnabled(d *caddyfile.Dispenser, metrics *map[string]Metric) error {
	if *metrics == nil {
		*metrics = map[string]Metric{}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		default:
			name := d.Val()
			metric := Metric{}
			if err := caddyutils.ParseBool(d, &metric.Enabled); err != nil {
				return err
			}
			(*metrics)[name] = metric
		}
	}
	return nil
}
