// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"errors"
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// ParseStringArrayMap parses a string array map from the dispenser.
// The following options can be used:
//   - AllowEmpty(): allows empty values. By default empty values are not allowed
//   - Default(): sets the default value of the string array map if no value is provided
//   - Inline(): parses the string array map as inline key-value pairs
//   - Separator(): sets the separator for inline key-value pairs
func ParseStringArrayMap(d *caddyfile.Dispenser, dest *map[string][]string, opts ...Option) error {
	p := &parseStringArrayMap{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	if p.inline && p.inlineSeparator == "" {
		return errors.New("inline separator and key-value separators must be set")
	}
	if p.inline && p.keyValueSeparator == "" {
		return errors.New("inline separator and key-value separators must be set")
	}
	return p.parse(d, dest)
}

type parseStringArrayMap struct {
	allowEmpty        bool
	allowEmptyValues  bool
	inline            bool
	inlineSeparator   string
	keyValueSeparator string
	defaultValue      *map[string][]string
}

func (p *parseStringArrayMap) SetAllowEmpty(b bool) {
	p.allowEmpty = b
}

func (p *parseStringArrayMap) SetAllowEmptyValues(b bool) {
	p.allowEmptyValues = b
}

func (p *parseStringArrayMap) SetInline(b bool) {
	p.inline = b
}

func (p *parseStringArrayMap) SetDefaultValue(value interface{}) error {
	v, ok := value.(map[string][]string)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseStringArrayMap) SetSeparator(s string, seps ...string) error {
	if len(seps) != 1 {
		return errors.New("expected two separators (inline separator and key-value separator)")
	}
	p.keyValueSeparator = s
	p.inlineSeparator = seps[0]
	return nil
}

func (p *parseStringArrayMap) parse(d *caddyfile.Dispenser, dest *map[string][]string) error {
	if p.inline {
		return p.parseInline(d, dest)
	} else {
		return p.parseBlock(d, dest)
	}
}

func (p *parseStringArrayMap) parseInline(d *caddyfile.Dispenser, dest *map[string][]string) error {
	var empty bool = true
	for d.NextArg() {
		raw := d.Val()
		var key string
		var values []string
		parts := strings.SplitN(raw, p.keyValueSeparator, 2)
		switch len(parts) {
		case 0:
			if !p.allowEmpty && p.defaultValue == nil {
				return d.Err("empty key values pair")
			}
		case 1:
			key = strings.TrimSpace(parts[0])
		default:
			key = strings.TrimSpace(parts[0])
			rawValues := strings.Split(parts[1], p.inlineSeparator)
			var emptyValues = true
			for _, val := range rawValues {
				val = strings.TrimSpace(val)
				if val == "" && !p.allowEmptyValues {
					continue
				}
				values = append(values, val)
				emptyValues = false
			}
			if emptyValues && !p.allowEmptyValues {
				return d.Err("expected at least one value")
			}
		}
		if len(values) == 0 && !p.allowEmptyValues {
			return d.Err("expected at least one value")
		}
		if len(values) == 0 {
			values = []string{}
		}
		if *dest == nil {
			*dest = make(map[string][]string)
		}
		if items, ok := (*dest)[key]; ok {
			(*dest)[key] = append(items, values...)
		} else {
			(*dest)[key] = values
		}
		empty = false
	}
	if empty && !p.allowEmpty {
		if p.defaultValue != nil && len(*p.defaultValue) > 0 {
			if *dest == nil {
				*dest = make(map[string][]string)
			}
			for key, values := range *p.defaultValue {
				(*dest)[key] = values
			}
		} else {
			return d.Err("empty block")
		}
	}
	return nil
}

func (p *parseStringArrayMap) parseBlock(d *caddyfile.Dispenser, dest *map[string][]string) error {
	var empty bool = true
	opts := []Option{}
	if p.allowEmptyValues {
		opts = append(opts, AllowEmpty(), AllowEmptyValues())
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		values := []string{}
		if err := ParseStringArray(d, &values, opts...); err != nil {
			return err
		}
		if *dest == nil {
			*dest = make(map[string][]string)
		}
		if items, ok := (*dest)[key]; ok {
			(*dest)[key] = append(items, values...)
		} else {
			(*dest)[key] = values
		}
		empty = false
	}
	if empty && !p.allowEmpty {
		if p.defaultValue != nil && len(*p.defaultValue) > 0 {
			if *dest == nil {
				*dest = make(map[string][]string)
			}
			for key, values := range *p.defaultValue {
				(*dest)[key] = values
			}
		} else {
			return d.Err("empty block")
		}
	}
	return nil
}

var (
	_ allowEmpty       = (*parseStringArrayMap)(nil)
	_ allowEmptyValues = (*parseStringArrayMap)(nil)
	_ allowInline      = (*parseStringArrayMap)(nil)
	_ allowDefault     = (*parseStringArrayMap)(nil)
	_ allowSeparator   = (*parseStringArrayMap)(nil)
)
