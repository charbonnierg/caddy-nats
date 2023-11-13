// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package filelog

import (
	"fmt"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func (r *FileLogReceiver) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d, parser.Match("filelog")); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "include":
			if err := parser.ParseStringArray(d, &r.Include); err != nil {
				return err
			}
		case "exclude":
			if err := parser.ParseStringArray(d, &r.Exclude); err != nil {
				return err
			}
		case "ordering_criteria":
			return d.Err("ordering_criteria is not supported yet")
		case "include_file_name":
			if err := parser.ParseBool(d, &r.IncludeFileName); err != nil {
				return err
			}
		case "include_file_path":
			if err := parser.ParseBool(d, &r.IncludeFilePath); err != nil {
				return err
			}
		case "include_file_name_resolved":
			if err := parser.ParseBool(d, &r.IncludeFileNameResolved); err != nil {
				return err
			}
		case "include_file_path_resolved":
			if err := parser.ParseBool(d, &r.IncludeFilePathResolved); err != nil {
				return err
			}
		case "poll_interval":
			if err := parser.ParseDuration(d, &r.PollInterval); err != nil {
				return err
			}
		case "start_at":
			if err := parser.ParseString(d, &r.StartAt); err != nil {
				return err
			}
		case "fingerprint_size":
			if err := parser.ParseInt64(d, &r.FingerprintSize); err != nil {
				return err
			}
		case "max_log_size":
			if err := parser.ParseInt64(d, &r.MaxLogSize); err != nil {
				return err
			}
		case "max_concurrent_files":
			if err := parser.ParseInt(d, &r.MaxConcurrentFiles); err != nil {
				return err
			}
		case "max_batches":
			if err := parser.ParseInt(d, &r.MaxBatches); err != nil {
				return err
			}
		case "delete_after_read":
			if err := parser.ParseBool(d, &r.DeleteAfterRead); err != nil {
				return err
			}
		case "multiline":
			return d.Err("multiline is not supported yet")
		case "preserve_leading_whitespaces":
			if err := parser.ParseBool(d, &r.PreserveLeading); err != nil {
				return err
			}
		case "preserve_trailing_whitespaces":
			if err := parser.ParseBool(d, &r.PreserveTrailing); err != nil {
				return err
			}
		case "encoding":
			if err := parser.ParseString(d, &r.Encoding); err != nil {
				return err
			}
		case "force_flush_period":
			if err := parser.ParseDuration(d, &r.FlushPeriod); err != nil {
				return err
			}
		case "attribute", "attributes":
			if err := parser.ParseStringMap(d, &r.Attributes); err != nil {
				return err
			}
		case "resource":
			if err := parser.ParseStringMap(d, &r.Resource); err != nil {
				return err
			}
		case "output", "outputs":
			if err := parser.ParseStringArray(d, &r.OutputIDs); err != nil {
				return err
			}
		case "operator":
			var name string
			if err := parser.ParseString(d, &name); err != nil {
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
			case "move":
				o := &Operator{
					MoveOperator: &MoveOperator{},
				}
				if err := ParseMoveOperator(d, o.MoveOperator); err != nil {
					return fmt.Errorf("move operator: %w", err)
				}
				r.Operators = append(r.Operators, o)
			case "remove":
				o := &Operator{
					RemoveOperator: &RemoveOperator{},
				}
				if err := ParseRemoveOperator(d, o.RemoveOperator); err != nil {
					return fmt.Errorf("remove operator: %w", err)
				}
				r.Operators = append(r.Operators, o)
			default:
				return d.Errf("unknown operator '%s'", name)
			}
		}
	}
	return nil
}

func ParseMoveOperator(d *caddyfile.Dispenser, r *MoveOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "from":
			if err := parser.ParseString(d, &r.From); err != nil {
				return err
			}
		case "to":
			if err := parser.ParseString(d, &r.To); err != nil {
				return err
			}
		case "output":
			if err := parser.ParseString(d, &r.Output); err != nil {
				return err
			}
		case "if":
			if err := parser.ParseString(d, &r.If); err != nil {
				return err
			}
		case "on_error":
			if err := parser.ParseString(d, &r.OnError); err != nil {
				return err
			}
		default:
			return d.Errf("unknown move operator property '%s'", d.Val())
		}
	}
	return nil
}

func ParseRemoveOperator(d *caddyfile.Dispenser, r *RemoveOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "field":
			if err := parser.ParseString(d, &r.Field); err != nil {
				return err
			}
		case "output":
			if err := parser.ParseString(d, &r.Output); err != nil {
				return err
			}
		case "if":
			if err := parser.ParseString(d, &r.If); err != nil {
				return err
			}
		case "on_error":
			if err := parser.ParseString(d, &r.OnError); err != nil {
				return err
			}
		default:
			return d.Errf("unknown remove operator property '%s'", d.Val())
		}
	}
	return nil
}

func ParseAddOperator(d *caddyfile.Dispenser, p *AddOperator) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "field":
			if err := parser.ParseString(d, &p.Field); err != nil {
				return err
			}
		case "value":
			if err := parser.ParseString(d, &p.Value); err != nil {
				return err
			}
		case "on_error":
			if err := parser.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "if":
			if err := parser.ParseString(d, &p.If); err != nil {
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
			if err := parser.ParseStringArray(d, &f.Include); err != nil {
				return err
			}
		case "output":
			if err := parser.ParseString(d, &f.Output); err != nil {
				return err
			}
		case "exclude":
			if err := parser.ParseStringArray(d, &f.Exclude); err != nil {
				return err
			}
		case "poll_interval":
			if err := parser.ParseDuration(d, &f.PollInterval); err != nil {
				return err
			}
		case "multiline":
			return d.Err("multiline is not supported yet")
		case "force_flush_period":
			if err := parser.ParseDuration(d, &f.ForceFlushPeriod); err != nil {
				return err
			}
		case "encoding":
			if err := parser.ParseString(d, &f.Encoding); err != nil {
				return err
			}
		case "include_file_name":
			if err := parser.ParseBool(d, &f.IncludeFileName); err != nil {
				return err
			}
		case "include_file_path":
			if err := parser.ParseBool(d, &f.IncludeFilePath); err != nil {
				return err
			}
		case "include_file_name_resolved":
			if err := parser.ParseBool(d, &f.IncludeFileNameResolved); err != nil {
				return err
			}
		case "include_file_path_resolved":
			if err := parser.ParseBool(d, &f.IncludeFilePathResolved); err != nil {
				return err
			}
		case "preserve_leading_whitespaces":
			if err := parser.ParseBool(d, &f.PreserveLeadingWhitespaces); err != nil {
				return err
			}
		case "preserve_trailing_whitespaces":
			if err := parser.ParseBool(d, &f.PreserveTrailingWhitespaces); err != nil {
				return err
			}
		case "start_at":
			if err := parser.ParseString(d, &f.StartAt); err != nil {
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
			if err := parser.ParseString(d, &p.Directory); err != nil {
				return err
			}
		case "files":
			if err := parser.ParseStringArray(d, &p.Files); err != nil {
				return err
			}
		case "units":
			if err := parser.ParseStringArray(d, &p.Units); err != nil {
				return err
			}
		case "match":
			p.Matches = fnutils.DefaultIfEmpty(p.Matches, []map[string]string{})
			m := map[string]string{}
			if err := parser.ParseStringMap(d, &m); err != nil {
				return err
			}
			p.Matches = append(p.Matches, m)
		case "matches":
			p.Matches = fnutils.DefaultIfEmpty(p.Matches, []map[string]string{})
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				m := map[string]string{}
				if err := parser.ParseStringMap(d, &m); err != nil {
					return err
				}
				p.Matches = append(p.Matches, m)
			}
		case "priority":
			if err := parser.ParseString(d, &p.Priority); err != nil {
				return err
			}
		case "grep":
			if err := parser.ParseString(d, &p.Grep); err != nil {
				return err
			}
		case "start_at":
			if err := parser.ParseString(d, &p.StartAt); err != nil {
				return err
			}
		case "attribute", "attributes":
			if err := parser.ParseStringMap(d, &p.Attributes); err != nil {
				return err
			}
		case "resource":
			if err := parser.ParseStringMap(d, &p.Resource); err != nil {
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
			if err := parser.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "layout_type":
			if err := parser.ParseString(d, &p.LayoutType); err != nil {
				return err
			}
		case "layout":
			if err := parser.ParseString(d, &p.Layout); err != nil {
				return err
			}
		case "location":
			if err := parser.ParseString(d, &p.Location); err != nil {
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
			if err := parser.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "on_error":
			if err := parser.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "preset":
			if err := parser.ParseString(d, &p.Preset); err != nil {
				return err
			}
		case "mapping":
			if err := parser.ParseStringArrayMap(d, &p.Mapping); err != nil {
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
			if err := parser.ParseString(d, &p.ParseFrom); err != nil {
				return err
			}
		case "parse_to":
			if err := parser.ParseString(d, &p.ParseTo); err != nil {
				return err
			}
		case "on_error":
			if err := parser.ParseString(d, &p.OnError); err != nil {
				return err
			}
		case "if":
			if err := parser.ParseString(d, &p.If); err != nil {
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
