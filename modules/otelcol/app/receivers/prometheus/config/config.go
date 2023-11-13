// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/alecthomas/units"
	"github.com/caddyserver/caddy/v2"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/model/labels"
)

type HTTPClientConfig promconfig.HTTPClientConfig

// GlobalConfig configures values that are used across other configuration
// objects.
type GlobalConfig struct {
	// How frequently to scrape targets by default.
	ScrapeInterval model.Duration `json:"scrape_interval,omitempty"`
	// The default timeout when scraping targets.
	ScrapeTimeout model.Duration `json:"scrape_timeout,omitempty"`
	// How frequently to evaluate rules by default.
	EvaluationInterval model.Duration `json:"evaluation_interval,omitempty"`
	// File to which PromQL queries are logged.
	QueryLogFile string `json:"query_log_file,omitempty"`
	// The labels to add to any timeseries that this Prometheus instance scrapes.
	ExternalLabels promlabels.Labels `json:"external_labels,omitempty"`
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
	// Keep no more than this many dropped targets per job.
	// 0 means no limit.
	KeepDroppedTargets uint `json:"keep_dropped_targets,omitempty"`
}

// Config is the top-level configuration for Prometheus's config files.
type Config struct {
	GlobalConfig      GlobalConfig    `json:"global"`
	ScrapeConfigFiles []string        `json:"scrape_config_files,omitempty"`
	ScrapeConfigs     []*ScrapeConfig `json:"scrape_configs,omitempty"`
}

func (c *Config) ReplaceAll(repl *caddy.Replacer) error {
	if c.GlobalConfig.ExternalLabels != nil {
		labels := promlabels.Labels{}
		for _, l := range c.GlobalConfig.ExternalLabels {
			name, err := repl.ReplaceOrErr(l.Name, true, true)
			if err != nil {
				return err
			}
			value, err := repl.ReplaceOrErr(l.Value, true, true)
			if err != nil {
				return err
			}
			labels = append(labels, promlabels.Label{Name: name, Value: value})
		}
		c.GlobalConfig.ExternalLabels = labels
	}
	if c.ScrapeConfigFiles != nil {
		for i, file := range c.ScrapeConfigFiles {
			file, err := repl.ReplaceOrErr(file, true, true)
			if err != nil {
				return err
			}
			c.ScrapeConfigFiles[i] = file
		}
	}
	for _, sc := range c.ScrapeConfigs {
		if err := sc.ReplaceAll(repl); err != nil {
			return err
		}
	}
	return nil
}
