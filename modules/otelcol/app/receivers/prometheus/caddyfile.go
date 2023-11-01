package prometheus

import (
	"net/url"
	"time"

	"github.com/alecthomas/units"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/model/labels"
	pconfig "github.com/quara-dev/beyond/modules/otelcol/app/receivers/prometheus/config"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (p *PrometheusReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "prometheus"); err != nil {
		return err
	}
	p.Config = fnutils.DefaultIfNil(p.Config, &pconfig.Config{})
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "scrape_interval":
			if err := parseModelDuration(d, &p.Config.GlobalConfig.ScrapeInterval); err != nil {
				return err
			}
		case "scrape_timeout":
			if err := parseModelDuration(d, &p.Config.GlobalConfig.ScrapeTimeout); err != nil {
				return err
			}
		case "evaluation_interval":
			if err := parseModelDuration(d, &p.Config.GlobalConfig.EvaluationInterval); err != nil {
				return err
			}
		case "query_log_file":
			if err := caddyutils.ParseString(d, &p.Config.GlobalConfig.QueryLogFile); err != nil {
				return err
			}
		case "external_labels":
			p.Config.GlobalConfig.ExternalLabels = fnutils.DefaultIfEmpty(p.Config.GlobalConfig.ExternalLabels, promlabels.Labels{})
			labels := map[string]string{}
			if err := caddyutils.ParseStringMap(d, &labels); err != nil {
				return err
			}
			for k, v := range labels {
				p.Config.GlobalConfig.ExternalLabels = append(p.Config.GlobalConfig.ExternalLabels, promlabels.Label{
					Name:  k,
					Value: v,
				})
			}
		case "body_size_limit":
			if err := parseBase2Bytes(d, &p.Config.GlobalConfig.BodySizeLimit); err != nil {
				return err
			}
		case "sample_limit":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.SampleLimit); err != nil {
				return err
			}
		case "target_limit":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.TargetLimit); err != nil {
				return err
			}
		case "label_limit":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.LabelLimit); err != nil {
				return err
			}
		case "label_name_lengh_limit":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.LabelNameLengthLimit); err != nil {
				return err
			}
		case "label_value_lengh_limit":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.LabelValueLengthLimit); err != nil {
				return err
			}
		case "keep_dropped_targets":
			if err := caddyutils.ParseUInt(d, &p.Config.GlobalConfig.KeepDroppedTargets); err != nil {
				return err
			}
		case "scrape_config":
			p.Config.ScrapeConfigs = fnutils.DefaultIfEmpty(p.Config.ScrapeConfigs, []*pconfig.ScrapeConfig{})
			if err := parseScrapeConfig(d, &p.Config.ScrapeConfigs); err != nil {
				return err
			}
		case "scrape_config_file":
			p.Config.ScrapeConfigFiles = fnutils.DefaultIfEmpty(p.Config.ScrapeConfigFiles, []string{})
			if err := caddyutils.ParseStringArray(d, &p.Config.ScrapeConfigFiles, false); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}

func parseScrapeConfig(d *caddyfile.Dispenser, cfgs *[]*pconfig.ScrapeConfig) error {
	if cfgs == nil {
		return d.Err("internal error: cfgs is nil")
	}
	cfg := &pconfig.ScrapeConfig{}
	if d.CountRemainingArgs() > 0 {
		cfg.StaticConfig = fnutils.DefaultIfEmpty(cfg.StaticConfig, []*pconfig.StaticConfig{})
		staticCfg := &pconfig.StaticConfig{
			Targets: []string{},
		}
		err := caddyutils.ParseString(d, &cfg.JobName)
		if err != nil {
			return err
		}
		if cfg.JobName == "" {
			return d.Err("job_name must be set")
		}
		err = caddyutils.ParseStringArray(d, &staticCfg.Targets, true)
		if err != nil {
			return err
		}
		cfg.StaticConfig = append(cfg.StaticConfig, staticCfg)
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "job_name":
			if err := caddyutils.ParseString(d, &cfg.JobName); err != nil {
				return err
			}
		case "honor_labels":
			if err := caddyutils.ParseBool(d, &cfg.HonorLabels); err != nil {
				return err
			}
		case "honor_timestamps":
			if err := caddyutils.ParseBool(d, &cfg.HonorTimestamps); err != nil {
				return err
			}
		case "params":
			cfg.Params = fnutils.DefaultIfEmptyMap(cfg.Params, url.Values{})
			var values map[string][]string
			if err := caddyutils.ParseStringArrayMap(d, &values); err != nil {
				return err
			}
			for k, v := range values {
				for _, item := range v {
					cfg.Params.Add(k, item)
				}
			}
		case "scrape_interval":
			if err := parseModelDuration(d, &cfg.ScrapeInterval); err != nil {
				return err
			}
		case "scrape_timeout":
			if err := parseModelDuration(d, &cfg.ScrapeTimeout); err != nil {
				return err
			}
		case "scrape_classic_histograms":
			if err := caddyutils.ParseBool(d, &cfg.ScrapeClassicHistograms); err != nil {
				return err
			}
		case "metrics_path":
			if err := caddyutils.ParseString(d, &cfg.MetricsPath); err != nil {
				return err
			}
		case "scheme":
			if err := caddyutils.ParseString(d, &cfg.Scheme); err != nil {
				return err
			}
		case "body_size_limit":
			if err := parseBase2Bytes(d, &cfg.BodySizeLimit); err != nil {
				return err
			}
		case "target_limit":
			if err := caddyutils.ParseUInt(d, &cfg.TargetLimit); err != nil {
				return err
			}
		case "label_limit":
			if err := caddyutils.ParseUInt(d, &cfg.LabelLimit); err != nil {
				return err
			}
		case "label_name_lengh_limit":
			if err := caddyutils.ParseUInt(d, &cfg.LabelNameLengthLimit); err != nil {
				return err
			}
		case "label_value_lengh_limit":
			if err := caddyutils.ParseUInt(d, &cfg.LabelValueLengthLimit); err != nil {
				return err
			}
		case "native_histogram_bucket_limit":
			if err := caddyutils.ParseUInt(d, &cfg.NativeHistogramBucketLimit); err != nil {
				return err
			}
		case "keep_dropped_targets":
			if err := caddyutils.ParseUInt(d, &cfg.KeepDroppedTargets); err != nil {
				return err
			}
		case "basic_auth":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			cfg.HTTPClientConfig.BasicAuth = fnutils.DefaultIfNil(cfg.HTTPClientConfig.BasicAuth, &config.BasicAuth{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "username":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.BasicAuth.Username); err != nil {
						return err
					}
				case "password":
					var secret string
					if err := caddyutils.ParseString(d, &secret); err != nil {
						return err
					}
					cfg.HTTPClientConfig.BasicAuth.Password = config.Secret(secret)
				case "password_file":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.BasicAuth.PasswordFile); err != nil {
						return err
					}
				}
			}
		case "bearer_token":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			var secret string
			if err := caddyutils.ParseString(d, &secret); err != nil {
				return err
			}
			cfg.HTTPClientConfig.BearerToken = config.Secret(secret)
		case "bearer_token_file":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.BearerTokenFile); err != nil {
				return err
			}
		case "authorization":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			cfg.HTTPClientConfig.Authorization = fnutils.DefaultIfNil(cfg.HTTPClientConfig.Authorization, &config.Authorization{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "type":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.Authorization.Type); err != nil {
						return err
					}
				case "credentials":
					var secret string
					if err := caddyutils.ParseString(d, &secret); err != nil {
						return err
					}
					cfg.HTTPClientConfig.Authorization.Credentials = config.Secret(secret)
				case "credentials_file":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.Authorization.CredentialsFile); err != nil {
						return err
					}
				}
			}
		case "oauth2":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			cfg.HTTPClientConfig.OAuth2 = fnutils.DefaultIfNil(cfg.HTTPClientConfig.OAuth2, &config.OAuth2{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "client_id":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.OAuth2.ClientID); err != nil {
						return err
					}
				case "client_secret":
					var secret string
					if err := caddyutils.ParseString(d, &secret); err != nil {
						return err
					}
					cfg.HTTPClientConfig.OAuth2.ClientSecret = config.Secret(secret)
				case "client_secret_file":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.OAuth2.ClientSecretFile); err != nil {
						return err
					}
				case "scope", "scopes":
					cfg.HTTPClientConfig.OAuth2.Scopes = fnutils.DefaultIfEmpty(cfg.HTTPClientConfig.OAuth2.Scopes, []string{})
					if err := caddyutils.ParseStringArray(d, &cfg.HTTPClientConfig.OAuth2.Scopes, false); err != nil {
						return err
					}
				case "token_url":
					if err := caddyutils.ParseString(d, &cfg.HTTPClientConfig.OAuth2.TokenURL); err != nil {
						return err
					}
				case "endpoint_parans":
					cfg.HTTPClientConfig.OAuth2.EndpointParams = fnutils.DefaultIfEmptyMap(cfg.HTTPClientConfig.OAuth2.EndpointParams, map[string]string{})
					if err := caddyutils.ParseStringMap(d, &cfg.HTTPClientConfig.OAuth2.EndpointParams); err != nil {
						return err
					}
				default:
					return d.Errf("unrecognized oauth2 http client option %s", d.Val())
				}
			}
		case "follow_redirects":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			if err := caddyutils.ParseBool(d, &cfg.HTTPClientConfig.FollowRedirects); err != nil {
				return err
			}
		case "enable_http2":
			cfg.HTTPClientConfig = fnutils.DefaultIfNil(cfg.HTTPClientConfig, &pconfig.HTTPClientConfig{})
			if err := caddyutils.ParseBool(d, &cfg.HTTPClientConfig.EnableHTTP2); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	*cfgs = append(*cfgs, cfg)
	return nil
}

func parseModelDuration(d *caddyfile.Dispenser, dest *model.Duration) error {
	var dur time.Duration
	if err := caddyutils.ParseDuration(d, &dur); err != nil {
		return err
	}
	*dest = model.Duration(dur)
	return nil
}

func parseBase2Bytes(d *caddyfile.Dispenser, dest *units.Base2Bytes) error {
	if !d.NextArg() {
		return d.ArgErr()
	}
	val, err := units.ParseBase2Bytes(d.Val())
	if err != nil {
		return err
	}
	*dest = val
	return nil
}
