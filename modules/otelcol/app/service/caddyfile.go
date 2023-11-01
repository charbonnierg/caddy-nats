package service

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
)

func (s *Service) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "extensions":
			if err := caddyutils.ParseStringArray(d, &s.Extensions, false); err != nil {
				return err
			}
		case "trace_pipeline":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "receivers":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Traces.Receivers, false); err != nil {
						return err
					}
				case "processors":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Traces.Processors, false); err != nil {
						return err
					}
				case "exporters":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Traces.Exporters, false); err != nil {
						return err
					}
				default:
					return d.Errf("unrecognized subdirective %s", d.Val())
				}
			}
		case "metric_pipeline":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "receivers":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Metrics.Receivers, false); err != nil {
						return err
					}
				case "processors":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Metrics.Processors, false); err != nil {
						return err
					}
				case "exporters":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Metrics.Exporters, false); err != nil {
						return err
					}
				default:
					return d.Errf("unrecognized subdirective %s", d.Val())
				}
			}
		case "log_pipeline":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "receivers":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Logs.Receivers, false); err != nil {
						return err
					}
				case "processors":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Logs.Processors, false); err != nil {
						return err
					}
				case "exporters":
					if err := caddyutils.ParseStringArray(d, &s.Pipelines.Logs.Exporters, false); err != nil {
						return err
					}
				default:
					return d.Errf("unrecognized subdirective %s", d.Val())
				}
			}
		case "log_level":
			if err := caddyutils.ParseString(d, &s.Telemetry.Logs.Level); err != nil {
				return err
			}
		case "log_extra_fields":
			if err := caddyutils.ParseStringMap(d, &s.Telemetry.Logs.InitialFields); err != nil {
				return err
			}
		case "metrics_level":
			if err := caddyutils.ParseString(d, &s.Telemetry.Metrics.Level); err != nil {
				return err
			}
		case "metrics":
			if err := caddyutils.ParseString(d, &s.Telemetry.Metrics.Address); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
