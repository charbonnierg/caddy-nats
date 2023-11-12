// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package filelog

import (
	"encoding/json"
	"errors"
	"time"
)

type Operator struct {
	Type string `json:"type,omitempty"`
	*FileInputOperator
	*JournaldInputOperator
	*JsonParserOperator
	*RegexParserOperator
	*TimeParser
	*AddOperator
	*FilterOperator
	*SeverityParser
	*RemoveOperator
}

func (o *Operator) MarshalJSON() ([]byte, error) {
	switch {
	case o.FileInputOperator != nil:
		o.Type = "file_input"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*FileInputOperator
		}{Type: o.Type, FileInputOperator: o.FileInputOperator})
	case o.JournaldInputOperator != nil:
		o.Type = "journald_input"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*JournaldInputOperator
		}{Type: o.Type, JournaldInputOperator: o.JournaldInputOperator})
	case o.JsonParserOperator != nil:
		o.Type = "json_parser"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*JsonParserOperator
		}{Type: o.Type, JsonParserOperator: o.JsonParserOperator})
	case o.RegexParserOperator != nil:
		o.Type = "regex_parser"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*RegexParserOperator
		}{Type: o.Type, RegexParserOperator: o.RegexParserOperator})
	case o.TimeParser != nil:
		o.Type = "time_parser"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*TimeParser
		}{Type: o.Type, TimeParser: o.TimeParser})
	case o.SeverityParser != nil:
		o.Type = "severity_parser"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*SeverityParser
		}{Type: o.Type, SeverityParser: o.SeverityParser})
	case o.AddOperator != nil:
		o.Type = "add"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*AddOperator
		}{Type: o.Type, AddOperator: o.AddOperator})
	case o.FilterOperator != nil:
		o.Type = "filter"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*FilterOperator
		}{Type: o.Type, FilterOperator: o.FilterOperator})
	case o.RemoveOperator != nil:
		o.Type = "remove"
		return json.Marshal(&struct {
			Type string `json:"type,omitempty"`
			*RemoveOperator
		}{Type: o.Type, RemoveOperator: o.RemoveOperator})
	default:
		return nil, errors.New("empty operator")
	}
}

type FileInputOperator struct {
	Include                     []string          `json:"include"`
	Output                      string            `json:"output,omitempty"`
	Exclude                     []string          `json:"exclude,omitempty"`
	PollInterval                time.Duration     `json:"poll_interval,omitempty"`
	Multiline                   *Split            `json:"multiline,omitempty"`
	ForceFlushPeriod            time.Duration     `json:"force_flush_period,omitempty"`
	Encoding                    string            `json:"encoding,omitempty"`
	IncludeFileName             bool              `json:"include_file_name,omitempty"`
	IncludeFilePath             bool              `json:"include_file_path,omitempty"`
	IncludeFileNameResolved     bool              `json:"include_file_name_resolved,omitempty"`
	IncludeFilePathResolved     bool              `json:"include_file_path_resolved,omitempty"`
	PreserveLeadingWhitespaces  bool              `json:"preserve_leading_whitespaces,omitempty"`
	PreserveTrailingWhitespaces bool              `json:"preserve_trailing_whitespaces,omitempty"`
	StartAt                     string            `json:"start_at,omitempty"`
	FingerprintSize             int               `json:"fingerprint_size,omitempty"`
	MaxLogSize                  int               `json:"max_log_size,omitempty"`
	MaxConcurrentFiles          int               `json:"max_concurrent_files,omitempty"`
	MaxBatches                  int               `json:"max_batches,omitempty"`
	DeleteAfterRead             bool              `json:"delete_after_read,omitempty"`
	Attributes                  map[string]string `json:"attributes,omitempty"`
	Resource                    map[string]string `json:"resource,omitempty"`
}

type JournaldInputOperator struct {
	Output     string              `json:"output,omitempty"`
	Directory  string              `json:"directory,omitempty"`
	Files      []string            `json:"files,omitempty"`
	Units      []string            `json:"units,omitempty"`
	Matches    []map[string]string `json:"matches,omitempty"`
	Priority   string              `json:"priority,omitempty"`
	Grep       string              `json:"grep,omitempty"`
	StartAt    string              `json:"start_at,omitempty"`
	Attributes map[string]string   `json:"attributes,omitempty"`
	Resource   map[string]string   `json:"resource,omitempty"`
}

type TimeParser struct {
	ParseFrom  string `json:"parse_from,omitempty"`
	LayoutType string `json:"layout_type,omitempty"`
	Layout     string `json:"layout,omitempty"`
	Location   string `json:"location,omitempty"`
}

type JsonParserOperator struct {
	Output    string          `json:"output,omitempty"`
	ParseFrom string          `json:"parse_from,omitempty"`
	ParseTo   string          `json:"parse_to,omitempty"`
	OnError   string          `json:"on_error,omitempty"`
	If        string          `json:"if,omitempty"`
	Timestamp *TimeParser     `json:"timestamp,omitempty"`
	Severity  *SeverityParser `json:"severity,omitempty"`
}

type RegexParserOperator struct {
	Output    string          `json:"output,omitempty"`
	Regex     string          `json:"regex,omitempty"`
	ParseFrom string          `json:"parse_from,omitempty"`
	ParseTo   string          `json:"parse_to,omitempty"`
	OnError   string          `json:"on_error,omitempty"`
	If        string          `json:"if,omitempty"`
	Timestamp *TimeParser     `json:"timestamp,omitempty"`
	Severity  *SeverityParser `json:"severity,omitempty"`
}

type SeverityParser struct {
	ParseFrom string              `json:"parse_from,omitempty"`
	OnError   string              `json:"on_error,omitempty"`
	Preset    string              `json:"preset,omitempty"`
	Mapping   map[string][]string `json:"mapping,omitempty"`
	If        string              `json:"if,omitempty"`
}

type AddOperator struct {
	Output  string `json:"output,omitempty"`
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	OnError string `json:"on_error,omitempty"`
	If      string `json:"if,omitempty"`
}

type FilterOperator struct {
	Output    string  `json:"output,omitempty"`
	Expr      string  `json:"expr,omitempty"`
	DropRatio float64 `json:"drop_ratio,omitempty"`
}

type RemoveOperator struct {
	Output  string `json:"output,omitempty"`
	Field   string `json:"field,omitempty"`
	OnError string `json:"on_error,omitempty"`
	If      string `json:"if,omitempty"`
}
