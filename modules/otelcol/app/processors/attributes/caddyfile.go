package attributes

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (r *AttributeProcessor) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "attributes"); err != nil {
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
					if err := caddyutils.ParseString(d, &filter.MatchType); err != nil {
						return err
					}
				case "service", "services":
					filter.Services = fnutils.DefaultIfEmpty(filter.Services, []string{})
					if err := caddyutils.ParseStringArray(d, &filter.Services, false); err != nil {
						return err
					}
				case "span_kind", "span_kinds":
					filter.SpanKinds = fnutils.DefaultIfEmpty(filter.SpanKinds, []string{})
					if err := caddyutils.ParseStringArray(d, &filter.SpanKinds, false); err != nil {
						return err
					}
				case "span_name", "span_names":
					filter.SpanNames = fnutils.DefaultIfEmpty(filter.SpanNames, []string{})
					if err := caddyutils.ParseStringArray(d, &filter.SpanNames, false); err != nil {
						return err
					}
				case "log_bodies", "log_body":
					if err := caddyutils.ParseStringArray(d, &filter.LogBodies, false); err != nil {
						return err
					}
				case "log_severity_text", "log_severity_texts":
					if err := caddyutils.ParseStringArray(d, &filter.LogSeverityTexts, false); err != nil {
						return err
					}
				case "log_severity_number", "log_severity_numbers":
					filter.LogSeverityNumber = &LogSeverityQuery{}
					for nesting := d.Nesting(); d.NextBlock(nesting); {
						switch d.Val() {
						case "min":
							if err := caddyutils.ParseInt(d, &filter.LogSeverityNumber.Min); err != nil {
								return err
							}
						case "match_undefined":
							if err := caddyutils.ParseBool(d, &filter.LogSeverityNumber.MatchUndefined); err != nil {
								return err
							}
						default:
							return d.Errf("unrecognized subdirective %s", d.Val())
						}
					}
				case "metric_name", "metric_names":
					filter.MetricNames = fnutils.DefaultIfEmpty(filter.MetricNames, []string{})
					if err := caddyutils.ParseStringArray(d, &filter.MetricNames, false); err != nil {
						return err
					}
				case "resource":
					filter.Resources = fnutils.DefaultIfEmpty(filter.Resources, []AttributeQuery{})
					var key string
					var value string
					if err := caddyutils.ParseString(d, &key); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &value); err != nil {
						return err
					}
					filter.Resources = append(filter.Resources, AttributeQuery{Key: key, Value: value})
				case "library":
					filter.Libraries = fnutils.DefaultIfEmpty(filter.Libraries, []LibraryQuery{})
					var name string
					var version string
					if err := caddyutils.ParseString(d, &name); err != nil {
						return err
					}
					if err := caddyutils.ParseString(d, &version); err != nil {
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
			if err := caddyutils.ParseString(d, &actionName); err != nil {
				return err
			}
			switch actionName {
			case "convert", "extract", "hash", "delete", "upsert", "update", "insert":
				r.Actions = fnutils.DefaultIfEmpty(r.Actions, []*Action{})
				action := Action{Action: actionName}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "key":
						if err := caddyutils.ParseString(d, &action.Key); err != nil {
							return err
						}
					case "from_attribute":
						if err := caddyutils.ParseString(d, &action.FromAttribute); err != nil {
							return err
						}
					case "from_context":
						if err := caddyutils.ParseString(d, &action.FromContext); err != nil {
							return err
						}
					case "pattern":
						if err := caddyutils.ParseString(d, &action.Pattern); err != nil {
							return err
						}
					case "converted_type":
						if err := caddyutils.ParseString(d, &action.ConvertedType); err != nil {
							return err
						}
					case "value":
						if err := caddyutils.ParseString(d, &action.Value); err != nil {
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
