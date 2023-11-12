// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"errors"
	"slices"
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseStringArray(d *caddyfile.Dispenser, dest *[]string, opts ...Option) error {
	p := &parseStringArray{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseStringArray struct {
	reverse          bool
	defaultValue     *[]string
	allowEmptyValues bool
	allowEmpty       bool
	separator        string
}

func (p *parseStringArray) SetReversed(reverse bool) {
	p.reverse = reverse
}

func (p *parseStringArray) SetDefaultValue(value any) error {
	v, ok := value.([]string)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseStringArray) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseStringArray) SetAllowEmptyValues(allowEmptyValues bool) {
	p.allowEmptyValues = allowEmptyValues
}

func (p *parseStringArray) SetSeparator(sep string, seps ...string) error {
	if len(seps) > 0 {
		return errors.New("expected a single separator")
	}
	p.separator = sep
	return nil
}

func (p *parseStringArray) parse(d *caddyfile.Dispenser, dest *[]string) error {
	var vals []string
	if p.separator != "" {
		vals = p.parseInline(d)
	} else {
		vals = p.parseArgs(d)
	}
	if len(vals) == 0 {
		switch {
		case p.defaultValue != nil:
			*dest = *p.defaultValue
		case p.allowEmpty:
			return nil
		default:
			return d.Err("expected at least one string value")
		}
	}
	if *dest == nil {
		*dest = []string{}
	}
	if p.reverse {
		slices.Reverse(vals)
	}
	*dest = append(*dest, vals...)
	return nil
}

func (p *parseStringArray) parseArgs(d *caddyfile.Dispenser) []string {
	vals := []string{}
	for d.NextArg() {
		val := d.Val()
		if val == "" && !p.allowEmptyValues {
			continue
		}
		vals = append(vals, val)
	}
	return vals
}

func (p *parseStringArray) parseInline(d *caddyfile.Dispenser) []string {
	vals := []string{}
	if !d.NextArg() {
		return nil
	}
	val := d.Val()
	if val == "" && !p.allowEmptyValues {
		return nil
	}
	parts := strings.Split(val, p.separator)
	for _, part := range parts {
		part := strings.TrimSpace(part)
		if part == "" && !p.allowEmptyValues {
			continue
		}
		vals = append(vals, part)
	}
	return vals
}

var (
	_ allowDefault   = (*parseStringArray)(nil)
	_ allowReverse   = (*parseStringArray)(nil)
	_ allowEmpty     = (*parseStringArray)(nil)
	_ allowSeparator = (*parseStringArray)(nil)
)
