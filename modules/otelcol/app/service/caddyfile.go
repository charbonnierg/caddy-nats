package service

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
)

func (s *Service) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "extensions":
			if err := parser.ParseStringArray(d, &s.Extensions); err != nil {
				return err
			}
		case "trace_pipeline":
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "receivers":
					if err := parser.ParseStringArray(d, &s.Pipelines.Traces.Receivers); err != nil {
						return err
					}
				case "processors":
					if err := parser.ParseStringArray(d, &s.Pipelines.Traces.Processors); err != nil {
						return err
					}
				case "exporters":
					if err := parser.ParseStringArray(d, &s.Pipelines.Traces.Exporters); err != nil {
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
					if err := parser.ParseStringArray(d, &s.Pipelines.Metrics.Receivers); err != nil {
						return err
					}
				case "processors":
					if err := parser.ParseStringArray(d, &s.Pipelines.Metrics.Processors); err != nil {
						return err
					}
				case "exporters":
					if err := parser.ParseStringArray(d, &s.Pipelines.Metrics.Exporters); err != nil {
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
					if err := parser.ParseStringArray(d, &s.Pipelines.Logs.Receivers); err != nil {
						return err
					}
				case "processors":
					if err := parser.ParseStringArray(d, &s.Pipelines.Logs.Processors); err != nil {
						return err
					}
				case "exporters":
					if err := parser.ParseStringArray(d, &s.Pipelines.Logs.Exporters); err != nil {
						return err
					}
				default:
					return d.Errf("unrecognized subdirective %s", d.Val())
				}
			}
		case "log_level":
			if err := parser.ParseString(d, &s.Telemetry.Logs.Level); err != nil {
				return err
			}
		case "log_extra_fields":
			if err := parser.ParseStringMap(d, &s.Telemetry.Logs.InitialFields); err != nil {
				return err
			}
		case "metrics_level":
			if err := parser.ParseString(d, &s.Telemetry.Metrics.Level); err != nil {
				return err
			}
		case "metrics":
			if err := parser.ParseString(d, &s.Telemetry.Metrics.Address); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
