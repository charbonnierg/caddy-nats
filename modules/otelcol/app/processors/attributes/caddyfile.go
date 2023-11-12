// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.opentelemetry.io/collector/component"
)

func (r *AttributeProcessor) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	var name string
	if err := parser.ParseString(d, &name); err != nil {
		return err
	}
	var id component.ID
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		directive := d.Val()
		switch directive {
		case "include", "exclude":
			filter := &Filter{}
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "match_type":
					if err := parser.ParseString(d, &filter.MatchType); err != nil {
						return err
					}
				case "service", "services":
					if err := parser.ParseStringArray(d, &filter.Services); err != nil {
						return err
					}
				case "span_kind", "span_kinds":
					if err := parser.ParseStringArray(d, &filter.SpanKinds); err != nil {
						return err
					}
				case "span_name", "span_names":
					if err := parser.ParseStringArray(d, &filter.SpanNames); err != nil {
						return err
					}
				case "log_bodies", "log_body":
					if err := parser.ParseStringArray(d, &filter.LogBodies); err != nil {
						return err
					}
				case "log_severity_text", "log_severity_texts":
					if err := parser.ParseStringArray(d, &filter.LogSeverityTexts); err != nil {
						return err
					}
				case "log_severity_number", "log_severity_numbers":
					filter.LogSeverityNumber = &LogSeverityQuery{}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "min":
							if err := parser.ParseInt(d, &filter.LogSeverityNumber.Min); err != nil {
								return err
							}
						case "match_undefined":
							if err := parser.ParseBool(d, &filter.LogSeverityNumber.MatchUndefined); err != nil {
								return err
							}
						default:
							return d.Errf("unrecognized subdirective %s", d.Val())
						}
					}
				case "metric_name", "metric_names":
					if err := parser.ParseStringArray(d, &filter.MetricNames); err != nil {
						return err
					}
				case "resource":
					filter.Resources = fnutils.DefaultIfEmpty(filter.Resources, []AttributeQuery{})
					var key string
					var value string
					if err := parser.ParseString(d, &key); err != nil {
						return err
					}
					if err := parser.ParseString(d, &value); err != nil {
						return err
					}
					filter.Resources = append(filter.Resources, AttributeQuery{Key: key, Value: value})
				case "library":
					filter.Libraries = fnutils.DefaultIfEmpty(filter.Libraries, []LibraryQuery{})
					var name string
					var version string
					if err := parser.ParseString(d, &name); err != nil {
						return err
					}
					if err := parser.ParseString(d, &version); err != nil {
						return err
					}
					filter.Libraries = append(filter.Libraries, LibraryQuery{Name: name, Version: version})
				}
			}
			if directive == "include" {
				r.Include = filter
			} else {
				r.Exclude = filter
			}
		case "action":
			actionName := ""
			if err := parser.ParseString(d, &actionName); err != nil {
				return err
			}
			switch actionName {
			case "convert", "extract", "hash", "delete", "upsert", "update", "insert":
				r.Actions = fnutils.DefaultIfEmpty(r.Actions, []*Action{})
				action := Action{Action: actionName}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "key":
						if err := parser.ParseString(d, &action.Key); err != nil {
							return err
						}
					case "from_attribute":
						if err := parser.ParseString(d, &action.FromAttribute); err != nil {
							return err
						}
					case "from_context":
						if err := parser.ParseString(d, &action.FromContext); err != nil {
							return err
						}
					case "pattern":
						if err := parser.ParseString(d, &action.Pattern); err != nil {
							return err
						}
					case "converted_type":
						if err := parser.ParseString(d, &action.ConvertedType); err != nil {
							return err
						}
					case "value":
						if err := parser.ParseString(d, &action.Value); err != nil {
							return err
						}
					default:
						return d.Errf("unrecognized subdirective %s", d.Val())
					}
				}
				r.Actions = append(r.Actions, &action)
			default:
				return d.Errf("unrecognized action %s", d.Val())
			}
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}
