// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/url"

	"github.com/alecthomas/units"
	"github.com/caddyserver/caddy/v2"
	"github.com/prometheus/common/model"
)

type StaticConfig struct {
	Targets []string `json:"targets"`
}

// FileSDConfig is the configuration for file based discovery.
type FileSDConfig struct {
	Files           []string       `json:"files"`
	RefreshInterval model.Duration `json:"refresh_interval,omitempty"`
}

// ScrapeConfig configures a scraping unit for Prometheus.
type ScrapeConfig struct {
	// HTTP client config
	*HTTPClientConfig
	// The job name to which the job label is set by default.
	JobName string `json:"job_name"`
	// Indicator whether the scraped metrics should remain unmodified.
	HonorLabels bool `json:"honor_labels,omitempty"`
	// Indicator whether the scraped timestamps should be respected.
	HonorTimestamps bool `json:"honor_timestamps,omitempty"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `json:"params,omitempty"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval model.Duration `json:"scrape_interval,omitempty"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout model.Duration `json:"scrape_timeout,omitempty"`
	// Whether to scrape a classic histogram that is also exposed as a native histogram.
	ScrapeClassicHistograms bool `json:"scrape_classic_histograms,omitempty"`
	// The HTTP resource path on which to fetch metrics from targets.
	MetricsPath string `json:"metrics_path,omitempty"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `json:"scheme,omitempty"`
	// An uncompressed response body larger than this many bytes will cause the
	// scrape to fail. 0 means no limit.
	BodySizeLimit units.Base2Bytes `json:"body_size_limit,omitempty"`
	// More than this many samples post metric-relabeling will cause the scrape to
	// fail. 0 means no limit.
	SampleLimit uint `json:"sample_limit,omitempty"`
	// More than this many targets after the target relabeling will cause the
	// scrapes to fail. 0 means no limit.
	TargetLimit uint `json:"target_limit,omitempty"`
	// More than this many labels post metric-relabeling will cause the scrape to
	// fail. 0 means no limit.
	LabelLimit uint `json:"label_limit,omitempty"`
	// More than this label name length post metric-relabeling will cause the
	// scrape to fail. 0 means no limit.
	LabelNameLengthLimit uint `json:"label_name_length_limit,omitempty"`
	// More than this label value length post metric-relabeling will cause the
	// scrape to fail. 0 means no limit.
	LabelValueLengthLimit uint `json:"label_value_length_limit,omitempty"`
	// More than this many buckets in a native histogram will cause the scrape to
	// fail.
	NativeHistogramBucketLimit uint `json:"native_histogram_bucket_limit,omitempty"`
	// Keep no more than this many dropped targets per job.
	// 0 means no limit.
	KeepDroppedTargets uint `json:"keep_dropped_targets,omitempty"`
	// List of target relabel configurations.
	RelabelConfigs []*RelabelConfig `json:"relabel_configs,omitempty"`
	// List of metric relabel configurations.
	MetricRelabelConfigs []*RelabelConfig `json:"metric_relabel_configs,omitempty"`
	// Discovery configurations.
	StaticConfig        []*StaticConfig        `json:"static_configs,omitempty"`
	FileSDConfig        []*FileSDConfig        `json:"file_sd_configs,omitempty"`
	DockerSDConfig      []*DockerSDConfig      `json:"docker_sd_configs,omitempty"`
	DockerSwarmSDConfig []*DockerSwarmSDConfig `json:"docker_swarm_sd_configs,omitempty"`
}

func (c *ScrapeConfig) ReplaceAll(repl *caddy.Replacer) error {
	if c.JobName != "" {
		jobName, err := repl.ReplaceOrErr(c.JobName, true, true)
		if err != nil {
			return err
		}
		c.JobName = jobName
	}
	for _, params := range c.Params {
		for i, param := range params {
			param, err := repl.ReplaceOrErr(param, true, true)
			if err != nil {
				return err
			}
			params[i] = param
		}
	}
	if c.MetricsPath != "" {
		metricsPath, err := repl.ReplaceOrErr(c.MetricsPath, true, true)
		if err != nil {
			return err
		}
		c.MetricsPath = metricsPath
	}
	if c.Scheme != "" {
		scheme, err := repl.ReplaceOrErr(c.Scheme, true, true)
		if err != nil {
			return err
		}
		c.Scheme = scheme
	}
	for _, rc := range c.RelabelConfigs {
		if err := rc.ReplaceAll(repl); err != nil {
			return err
		}
	}
	for _, rc := range c.MetricRelabelConfigs {
		if err := rc.ReplaceAll(repl); err != nil {
			return err
		}
	}
	for _, sc := range c.StaticConfig {
		for i, target := range sc.Targets {
			target, err := repl.ReplaceOrErr(target, true, true)
			if err != nil {
				return err
			}
			sc.Targets[i] = target
		}
	}
	for _, fsc := range c.FileSDConfig {
		for i, file := range fsc.Files {
			file, err := repl.ReplaceOrErr(file, true, true)
			if err != nil {
				return err
			}
			fsc.Files[i] = file
		}
	}
	for _, dsc := range c.DockerSDConfig {
		if dsc.Host != "" {
			host, err := repl.ReplaceOrErr(dsc.Host, true, true)
			if err != nil {
				return err
			}
			dsc.Host = host
		}
		for _, filter := range dsc.Filters {
			for i, value := range filter.Values {
				value, err := repl.ReplaceOrErr(value, true, true)
				if err != nil {
					return err
				}
				filter.Values[i] = value
			}
		}
	}
	for _, dssc := range c.DockerSwarmSDConfig {
		if dssc.Host != "" {
			host, err := repl.ReplaceOrErr(dssc.Host, true, true)
			if err != nil {
				return err
			}
			dssc.Host = host
		}
		if dssc.Role != "" {
			role, err := repl.ReplaceOrErr(dssc.Role, true, true)
			if err != nil {
				return err
			}
			dssc.Role = role
		}
		for _, filter := range dssc.Filters {
			for i, value := range filter.Values {
				value, err := repl.ReplaceOrErr(value, true, true)
				if err != nil {
					return err
				}
				filter.Values[i] = value
			}
		}
	}
	return nil
}
