package filelog

import (
	"fmt"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (r *FileLogReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := caddyutils.ExpectString(d, "filelog"); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include":
			if err := caddyutils.ParseStringArray(d, &r.Include, false); err != nil {
				return err
			}
		case "exclude":
			if err := caddyutils.ParseStringArray(d, &r.Exclude, false); err != nil {
				return err
			}
		case "ordering_criteria":
			return d.Err("ordering_criteria is not supported yet")
		case "include_file_name":
			if err := caddyutils.ParseBool(d, &r.IncludeFileName); err != nil {
				return err
			}
		case "include_file_path":
			if err := caddyutils.ParseBool(d, &r.IncludeFilePath); err != nil {
				return err
			}
		case "include_file_name_resolved":
			if err := caddyutils.ParseBool(d, &r.IncludeFileNameResolved); err != nil {
				return err
			}
		case "include_file_path_resolved":
			if err := caddyutils.ParseBool(d, &r.IncludeFilePathResolved); err != nil {
				return err
			}
		case "poll_interval":
			if err := caddyutils.ParseDuration(d, &r.PollInterval); err != nil {
				return err
			}
		case "start_at":
			if err := caddyutils.ParseString(d, &r.StartAt); err != nil {
				return err
			}
		case "fingerprint_size":
			if err := caddyutils.ParseInt64(d, &r.FingerprintSize); err != nil {
				return err
			}
		case "max_log_size":
			if err := caddyutils.ParseInt64(d, &r.MaxLogSize); err != nil {
				return err
			}
		case "max_concurrent_files":
			if err := caddyutils.ParseInt(d, &r.MaxConcurrentFiles); err != nil {
				return err
			}
		case "max_batches":
			if err := caddyutils.ParseInt(d, &r.MaxBatches); err != nil {
				return err
			}
		case "delete_after_read":
			if err := caddyutils.ParseBool(d, &r.DeleteAfterRead); err != nil {
				return err
			}
		case "multiline":
			return d.Err("multiline is not supported yet")
		case "preserve_leading_whitespaces":
			if err := caddyutils.ParseBool(d, &r.PreserveLeading); err != nil {
				return err
			}
		case "preserve_trailing_whitespaces":
			if err := caddyutils.ParseBool(d, &r.PreserveTrailing); err != nil {
				return err
			}
		case "encoding":
			if err := caddyutils.ParseString(d, &r.Encoding); err != nil {
				return err
			}
		case "force_flush_period":
			if err := caddyutils.ParseDuration(d, &r.FlushPeriod); err != nil {
				return err
			}
		case "attribute", "attributes":
			if err := caddyutils.ParseStringMap(d, &r.Attributes); err != nil {
				return err
			}
		case "resource":
			if err := caddyutils.ParseStringMap(d, &r.Resource); err != nil {
				return err
			}
		case "output", "outputs":
			if err := caddyutils.ParseStringArray(d, &r.OutputIDs, false); err != nil {
				return err
			}
		case "operator":
			var name string
			if err := caddyutils.ParseString(d, &name); err != nil {
				return err
			}
			r.Operators = fnutils.DefaultIfEmpty(r.Operators, []*Operator{})
			switch name {
			case "file_input":
				o := &Operator{
					FileInputOperator: &FileInputOperator{},
				}
				if err := ParseFileInputOperator(d, o.FileInputOperator); err != nil {
					return err
				}
				r.Operators = append(r.Operators, o)
			case "journald_input":
				o := &Operator{
					JournaldInputOperator: &JournaldInputOperator{},
				}
				if err := ParseJournaldInputOperator(d, o.JournaldInputOperator); err != nil {
					return err
				}
				r.Operators = append(r.Operators, o)
			case "json_parser":
				o := &Operator{
					JsonParserOperator: &JsonParserOperator{},
				}
				if err := ParseJsonParserOperator(d, o.JsonParserOperator); err != nil {
					return err
				}
				r.Operators = append(r.Operators, o)
			case "regex_parser":
				return d.Err("regex_parser is not supported yet")
			case "severity_parser":
				o := &Operator{
					SeverityParser: &SeverityParser{},
				}
				if err := ParseSeverityParserOperator(d, o.SeverityParser); err != nil {
					return err
				}
				r.Operators = append(r.Operators, o)
			case "time_parser":
				o := &Operator{
					TimeParser: &TimeParser{},
				}
				if err := ParseTimeParserOperator(d, o.TimeParser); err != nil {
					return err
				}
				r.Operators = append(r.Operators, o)
			case "add":
				o := &Operator{
					AddOperator: &AddOperator{},
				}
				if err := ParseAddOperator(d, o.AddOperator); err != nil {
					return fmt.Errorf("add operator: %w", err)
				}
				r.Operators = append(r.Operators, o)
			case "filter":
				return d.Err("filter is not supported yet")
			case "remove":
				return d.Err("remove is not supported yet")
			}
		}
	}
	return nil
}

func ParseAddOperator(d *caddyfile.Dispenser, p *AddOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "field":
			if err := caddyutils.ParseString(d, &p.Field); err != nil {
				return err
			}
		case "value":
			if err := caddyutils.ParseString(d, &p.Value); err != nil {
				return err
			}
		case "on_error":
			if err := caddyutils.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "if":
			if err := caddyutils.ParseString(d, &p.If); err != nil {
				return err
			}
		default:
			return d.Errf("unknown add operator property '%s'", d.Val())
		}
	}
	return nil
}

func ParseFileInputOperator(d *caddyfile.Dispenser, f *FileInputOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include":
			if err := caddyutils.ParseStringArray(d, &f.Include, false); err != nil {
				return err
			}
		case "output":
			if err := caddyutils.ParseString(d, &f.Output); err != nil {
				return err
			}
		case "exclude":
			if err := caddyutils.ParseStringArray(d, &f.Exclude, false); err != nil {
				return err
			}
		case "poll_interval":
			if err := caddyutils.ParseDuration(d, &f.PollInterval); err != nil {
				return err
			}
		case "multiline":
			return d.Err("multiline is not supported yet")
		case "force_flush_period":
			if err := caddyutils.ParseDuration(d, &f.ForceFlushPeriod); err != nil {
				return err
			}
		case "encoding":
			if err := caddyutils.ParseString(d, &f.Encoding); err != nil {
				return err
			}
		case "include_file_name":
			if err := caddyutils.ParseBool(d, &f.IncludeFileName); err != nil {
				return err
			}
		case "include_file_path":
			if err := caddyutils.ParseBool(d, &f.IncludeFilePath); err != nil {
				return err
			}
		case "include_file_name_resolved":
			if err := caddyutils.ParseBool(d, &f.IncludeFileNameResolved); err != nil {
				return err
			}
		case "include_file_path_resolved":
			if err := caddyutils.ParseBool(d, &f.IncludeFilePathResolved); err != nil {
				return err
			}
		case "preserve_leading_whitespaces":
			if err := caddyutils.ParseBool(d, &f.PreserveLeadingWhitespaces); err != nil {
				return err
			}
		case "preserve_trailing_whitespaces":
			if err := caddyutils.ParseBool(d, &f.PreserveTrailingWhitespaces); err != nil {
				return err
			}
		case "start_at":
			if err := caddyutils.ParseString(d, &f.StartAt); err != nil {
				return err
			}
		default:
			return d.Errf("unknown property '%s'", d.Val())
		}
	}
	return nil
}

func ParseJournaldInputOperator(d *caddyfile.Dispenser, p *JournaldInputOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "directory":
			if err := caddyutils.ParseString(d, &p.Directory); err != nil {
				return err
			}
		case "files":
			if err := caddyutils.ParseStringArray(d, &p.Files, false); err != nil {
				return err
			}
		case "units":
			if err := caddyutils.ParseStringArray(d, &p.Units, false); err != nil {
				return err
			}
		case "match":
			p.Matches = fnutils.DefaultIfEmpty(p.Matches, []map[string]string{})
			m := map[string]string{}
			if err := caddyutils.ParseStringMap(d, &m); err != nil {
				return err
			}
			p.Matches = append(p.Matches, m)
		case "matches":
			p.Matches = fnutils.DefaultIfEmpty(p.Matches, []map[string]string{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				m := map[string]string{}
				if err := caddyutils.ParseStringMap(d, &m); err != nil {
					return err
				}
				p.Matches = append(p.Matches, m)
			}
		case "priority":
			if err := caddyutils.ParseString(d, &p.Priority); err != nil {
				return err
			}
		case "grep":
			if err := caddyutils.ParseString(d, &p.Grep); err != nil {
				return err
			}
		case "start_at":
			if err := caddyutils.ParseString(d, &p.StartAt); err != nil {
				return err
			}
		case "attribute", "attributes":
			if err := caddyutils.ParseStringMap(d, &p.Attributes); err != nil {
				return err
			}
		case "resource":
			if err := caddyutils.ParseStringMap(d, &p.Resource); err != nil {
				return err
			}
		default:
			return d.Errf("unknown journald_input operator property '%s'", d.Val())
		}
	}
	return nil
}

func ParseTimeParserOperator(d *caddyfile.Dispenser, p *TimeParser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "parse_from":
			if err := caddyutils.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "layout_type":
			if err := caddyutils.ParseString(d, &p.LayoutType); err != nil {
				return err
			}
		case "layout":
			if err := caddyutils.ParseString(d, &p.Layout); err != nil {
				return err
			}
		case "location":
			if err := caddyutils.ParseString(d, &p.Location); err != nil {
				return err
			}
		default:
			return d.Errf("unknown property '%s'", d.Val())
		}
	}
	return nil
}

func ParseSeverityParserOperator(d *caddyfile.Dispenser, p *SeverityParser) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "parse_from":
			if err := caddyutils.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "on_error":
			if err := caddyutils.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "preset":
			if err := caddyutils.ParseString(d, &p.Preset); err != nil {
				return err
			}
		case "mapping":
			if err := caddyutils.ParseStringArrayMap(d, &p.Mapping); err != nil {
				return err
			}
		default:
			return d.Errf("unknown property '%s'", d.Val())
		}
	}
	return nil
}

func ParseJsonParserOperator(d *caddyfile.Dispenser, p *JsonParserOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "parse_from":
			if err := caddyutils.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "parse_to":
			if err := caddyutils.ParseString(d, &p.ParseTo); err != nil {
				return err
			}
		case "on_error":
			if err := caddyutils.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "if":
			if err := caddyutils.ParseString(d, &p.If); err != nil {
				return err
			}
		case "timestamp":
			p.Timestamp = &TimeParser{}
			if err := ParseTimeParserOperator(d, p.Timestamp); err != nil {
				return err
			}
		case "severity":
			p.Severity = &SeverityParser{}
			if err := ParseSeverityParserOperator(d, p.Severity); err != nil {
				return err
			}
		default:
			return d.Errf("unknown json_parser operator property '%s'", d.Val())
		}
	}
	return nil
}
