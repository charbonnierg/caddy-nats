// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package filelog

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
	"github.com/quara-dev/beyond/modules/otelcol/app/settings"
)

func init() {
	caddy.RegisterModule(&FileLogReceiver{})
}

type OrderingCriteria struct {
	Regex  string `json:"regex,omitempty"`
	SortBy []Sort `json:"sort_by,omitempty"`
}

type Sort struct {
	SortType  string `json:"sort_type,omitempty"`
	RegexKey  string `json:"regex_key,omitempty"`
	Ascending bool   `json:"ascending,omitempty"`

	// Timestamp only
	Layout   string `json:"layout,omitempty"`
	Location string `json:"location,omitempty"`
}

type Split struct {
	LineStartPattern string `json:"line_start_pattern,omitempty"`
	LineEndPattern   string `json:"line_end_pattern,omitempty"`
	OmitPattern      bool   `json:"omit_pattern,omitempty"`
}

type FileLogReceiver struct {
	Include                 []string              `json:"include,omitempty"`
	Exclude                 []string              `json:"exclude,omitempty"`
	OrderingCriteria        *OrderingCriteria     `json:"ordering_criteria,omitempty"`
	IncludeFileName         bool                  `json:"include_file_name,omitempty"`
	IncludeFilePath         bool                  `json:"include_file_path,omitempty"`
	IncludeFileNameResolved bool                  `json:"include_file_name_resolved,omitempty"`
	IncludeFilePathResolved bool                  `json:"include_file_path_resolved,omitempty"`
	PollInterval            time.Duration         `json:"poll_interval,omitempty"`
	StartAt                 string                `json:"start_at,omitempty"`
	FingerprintSize         int64                 `json:"fingerprint_size,omitempty"`
	MaxLogSize              int64                 `json:"max_log_size,omitempty"`
	MaxConcurrentFiles      int                   `json:"max_concurrent_files,omitempty"`
	MaxBatches              int                   `json:"max_batches,omitempty"`
	DeleteAfterRead         bool                  `json:"delete_after_read,omitempty"`
	SplitConfig             *Split                `json:"multiline,omitempty"`
	PreserveLeading         bool                  `json:"preserve_leading_whitespaces,omitempty"`
	PreserveTrailing        bool                  `json:"preserve_trailing_whitespaces,omitempty"`
	Encoding                string                `json:"encoding,omitempty"`
	FlushPeriod             time.Duration         `json:"force_flush_period,omitempty"`
	Attributes              map[string]string     `json:"attributes,omitempty"`
	Resource                map[string]string     `json:"resource,omitempty"`
	OutputIDs               []string              `json:"output,omitempty"`
	OperatorID              string                `json:"id,omitempty"`
	OperatorType            string                `json:"type,omitempty"`
	Operators               []*Operator           `json:"operators,omitempty"`
	StorageID               string                `json:"storage,omitempty"`
	RetryOnFailure          *settings.RetryConfig `json:"retry_on_failure,omitempty"`
}

func (FileLogReceiver) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.receivers.filelog",
		New: func() caddy.Module { return new(FileLogReceiver) },
	}
}

func (f *FileLogReceiver) ReplaceAll(repl *caddy.Replacer) error {
	if f.OrderingCriteria != nil {
		if f.OrderingCriteria.Regex != "" {
			regex, err := repl.ReplaceOrErr(f.OrderingCriteria.Regex, true, true)
			if err != nil {
				return err
			}
			f.OrderingCriteria.Regex = regex
		}
		if f.OrderingCriteria.SortBy != nil {
			for _, sort := range f.OrderingCriteria.SortBy {
				if sort.RegexKey != "" {
					regexKey, err := repl.ReplaceOrErr(sort.RegexKey, true, true)
					if err != nil {
						return err
					}
					sort.RegexKey = regexKey
				}
				if sort.Layout != "" {
					layout, err := repl.ReplaceOrErr(sort.Layout, true, true)
					if err != nil {
						return err
					}
					sort.Layout = layout
				}
				if sort.Location != "" {
					location, err := repl.ReplaceOrErr(sort.Location, true, true)
					if err != nil {
						return err
					}
					sort.Location = location
				}
			}
		}
	}
	if f.SplitConfig != nil {
		if f.SplitConfig.LineStartPattern != "" {
			lineStartPattern, err := repl.ReplaceOrErr(f.SplitConfig.LineStartPattern, true, true)
			if err != nil {
				return err
			}
			f.SplitConfig.LineStartPattern = lineStartPattern
		}
		if f.SplitConfig.LineEndPattern != "" {
			lineEndPattern, err := repl.ReplaceOrErr(f.SplitConfig.LineEndPattern, true, true)
			if err != nil {
				return err
			}
			f.SplitConfig.LineEndPattern = lineEndPattern
		}
	}
	if f.StartAt != "" {
		startAt, err := repl.ReplaceOrErr(f.StartAt, true, true)
		if err != nil {
			return err
		}
		f.StartAt = startAt
	}
	if f.Encoding != "" {
		encoding, err := repl.ReplaceOrErr(f.Encoding, true, true)
		if err != nil {
			return err
		}
		f.Encoding = encoding
	}
	return nil
}

var (
	_ config.Receiver = (*FileLogReceiver)(nil)
)
