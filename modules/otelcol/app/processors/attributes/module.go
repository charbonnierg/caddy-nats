// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package attributes

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/otelcol/app/config"
)

func init() {
	caddy.RegisterModule(AttributeProcessor{})
}

type Action struct {
	Action        string `json:"action,omitempty"`
	Key           string `json:"key,omitempty"`
	Value         string `json:"value,omitempty"`
	FromAttribute string `json:"from_attribute,omitempty"`
	FromContext   string `json:"from_context,omitempty"`
	Pattern       string `json:"pattern,omitempty"`
	ConvertedType string `json:"converted_type,omitempty"`
}

type Regexp struct {
	CacheEnabled       bool `json:"cacheenabled,omitempty"`
	CacheMaxNumEntries int  `json:"cachemaxnumentries,omitempty"`
}

type AttributeQuery struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type LibraryQuery struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type LogSeverityQuery struct {
	Min            int  `json:"min,omitempty"`
	MatchUndefined bool `json:"match_undefined,omitempty"`
}

type Filter struct {
	MatchType         string            `json:"match_type,omitempty"`
	Regexp            *Regexp           `json:"regexp,omitempty"`
	Services          []string          `json:"services,omitempty"`
	Resources         []AttributeQuery  `json:"resources,omitempty"`
	Libraries         []LibraryQuery    `json:"libraries,omitempty"`
	SpanNames         []string          `json:"span_names,omitempty"`
	SpanKinds         []string          `json:"span_kinds,omitempty"`
	LogBodies         []string          `json:"log_bodies,omitempty"`
	LogSeverityTexts  []string          `json:"log_severity_texts,omitempty"`
	LogSeverityNumber *LogSeverityQuery `json:"log_severity_number,omitempty"`
	MetricNames       []string          `json:"metric_names,omitempty"`
	Attributes        []AttributeQuery  `json:"attributes,omitempty"`
}

type AttributeProcessor struct {
	Actions []*Action `json:"actions,omitempty"`
	Include *Filter   `json:"include,omitempty"`
	Exclude *Filter   `json:"exclude,omitempty"`
}

func (AttributeProcessor) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "otelcol.processors.attributes",
		New: func() caddy.Module { return new(AttributeProcessor) },
	}
}

// Interface guards
var (
	_ config.Processor = (*AttributeProcessor)(nil)
)
