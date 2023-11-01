package config

import (
	"net/url"

	"github.com/alecthomas/units"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
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
	RelabelConfigs []*relabel.Config `json:"relabel_configs,omitempty"`
	// List of metric relabel configurations.
	MetricRelabelConfigs []*relabel.Config `json:"metric_relabel_configs,omitempty"`
	// Discovery configurations.
	StaticConfig        []*StaticConfig        `json:"static_configs,omitempty"`
	FileSDConfig        []*FileSDConfig        `json:"file_sd_configs,omitempty"`
	DockerSDConfig      []*DockerSDConfig      `json:"docker_sd_configs,omitempty"`
	DockerSwarmSDConfig []*DockerSwarmSDConfig `json:"docker_swarm_sd_configs,omitempty"`
}
