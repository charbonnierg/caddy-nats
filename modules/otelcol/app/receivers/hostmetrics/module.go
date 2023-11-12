// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package hostmetrics

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
)

func init() {
	caddy.RegisterModule(HostMetricsReceiver{})
}

type Metric struct {
	Enabled bool `json:"enabled,omitempty"`
}

type CpuScraper struct {
	Metrics map[string]Metric `json:"metrics,omitempty"`
}

type DeviceFilter struct {
	DeviceNames []string `json:"devices,omitempty"`
	MatchType   string   `json:"match_type,omitempty"`
}

type FSTypeFilter struct {
	FSTypes   []string `json:"fs_types,omitempty"`
	MatchType string   `json:"match_type,omitempty"`
}

type MountPointFilter struct {
	MountPoints []string `json:"mount_points,omitempty"`
	MatchType   string   `json:"match_type,omitempty"`
}

type NetworkInterfaceFilter struct {
	Interfaces []string `json:"interfaces,omitempty"`
	MatchType  string   `json:"match_type,omitempty"`
}

type DiskScraper struct {
	Metrics            map[string]Metric `json:"metrics,omitempty"`
	IncludeDevices     *DeviceFilter     `json:"include_devices,omitempty"`
	ExcludeDevices     *DeviceFilter     `json:"exclude_devices,omitempty"`
	IncludeFSTypes     *FSTypeFilter     `json:"include_fs_types,omitempty"`
	ExcludeFSTypes     *FSTypeFilter     `json:"exclude_fs_types,omitempty"`
	IncludeMountPoints *MountPointFilter `json:"include_mount_points,omitempty"`
	ExcludeMountPoints *MountPointFilter `json:"exclude_mount_points,omitempty"`
}

type LoadScraper struct {
	Metrics    map[string]Metric `json:"metrics,omitempty"`
	CpuAverage bool              `json:"cpu_average,omitempty"`
}

type FilesystemScraper struct {
	Metrics map[string]Metric `json:"metrics,omitempty"`
}
type MemoryScrapper struct {
	Metrics map[string]Metric `json:"metrics,omitempty"`
}
type NetworkScrapper struct {
	Metrics           map[string]Metric       `json:"metrics,omitempty"`
	IncludeInterfaces *NetworkInterfaceFilter `json:"include,omitempty"`
	ExcludeInterfaces *NetworkInterfaceFilter `json:"exclude,omitempty"`
}
type PagingScraper struct {
	Metrics map[string]Metric `json:"metrics,omitempty"`
}
type ProcessesScraper struct {
	Metrics map[string]Metric `json:"metrics,omitempty"`
}

type ProcessNameFilter struct {
	Names     []string `json:"names,omitempty"`
	MatchType string   `json:"match_type,omitempty"`
}

type ProcessScapper struct {
	Metrics              map[string]Metric  `json:"metrics,omitempty"`
	IncludeProcesses     *ProcessNameFilter `json:"include,omitempty"`
	ExcludeProcesses     *ProcessNameFilter `json:"exclude,omitempty"`
	MuteProcessNameError bool               `json:"mute_process_name_error,omitempty"`
	MuteProcessExeError  bool               `json:"mute_process_exe_error,omitempty"`
	MuteProcessIOErrror  bool               `json:"mute_process_io_error,omitempty"`
	MuteProcessUserError bool               `json:"mute_process_user_error,omitempty"`
	ScrapeProcessDelay   time.Duration      `json:"scrape_process_delay,omitempty"`
}

type Scrapers struct {
	Cpu        *CpuScraper        `json:"cpu,omitempty"`
	Disk       *DiskScraper       `json:"disk,omitempty"`
	Load       *LoadScraper       `json:"load,omitempty"`
	Filesystem *FilesystemScraper `json:"filesystem,omitempty"`
	Memory     *MemoryScrapper    `json:"memory,omitempty"`
	Network    *NetworkScrapper   `json:"network,omitempty"`
	Paging     *PagingScraper     `json:"paging,omitempty"`
	Processes  *ProcessesScraper  `json:"processes,omitempty"`
	Process    *ProcessScapper    `json:"process,omitempty"`
}

type HostMetricsReceiver struct {
	// RootPath is the host's root directory (linux only).
	RootPath string `json:"root_path,omitempty"`
	// CollectionInterval sets the how frequently the scraper
	// should be called and used as the context timeout
	// to ensure that scrapers don't exceed the interval.
	CollectionInterval time.Duration `json:"collection_interval"`
	// InitialDelay sets the initial start delay for the scraper,
	// any non positive value is assumed to be immediately.
	InitialDelay time.Duration `json:"initial_delay"`
	// Timeout is an optional value used to set scraper's context deadline.
	Timeout  time.Duration `json:"timeout"`
	Scrapers *Scrapers     `json:"scrapers,omitempty"`
}

func (HostMetricsReceiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.receivers.hostmetrics",
		New: func() caddy.Module { return new(HostMetricsReceiver) },
	}
}

var (
	_ config.Receiver = (*HostMetricsReceiver)(nil)
)
