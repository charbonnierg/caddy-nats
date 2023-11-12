// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"errors"
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// ParseStringMap parses a string map from the dispenser.
// The following options can be used:
//   - AllowEmpty(): allows empty values. By default empty values are not allowed
//   - Default(): sets the default value of the string map if no value is provided
//   - Inline(): parses the string map as inline key-value pairs
//   - Separator(): sets the separator for inline key-value pairs
func ParseStringMap(d *caddyfile.Dispenser, dest *map[string]string, opts ...Option) error {
	p := &parseStringMap{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseStringMap struct {
	allowEmpty        bool
	defaultValue      *map[string]string
	inline            bool
	keyValueSeparator string
}

func (p *parseStringMap) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseStringMap) SetDefaultValue(value interface{}) error {
	v, ok := value.(map[string]string)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseStringMap) SetInline(value bool) {
	p.inline = value
}

func (p *parseStringMap) SetSeparator(value string, seps ...string) error {
	p.keyValueSeparator = value
	if len(seps) > 0 {
		return errors.New("expected a single separator")
	}
	return nil
}

func (p *parseStringMap) parse(d *caddyfile.Dispenser, dest *map[string]string) error {
	if p.inline {
		return p.parseInline(d, dest)
	}
	var empty = true
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		key := d.Val()
		if !d.NextArg() {
			return d.Err("expected a string value")
		}
		value := d.Val()
		if _, ok := (*dest)[key]; ok {
			return d.Err("duplicate key")
		}
		if *dest == nil {
			*dest = make(map[string]string)
		}
		(*dest)[key] = value
		empty = false
	}
	if empty && !p.allowEmpty {
		if p.defaultValue != nil {
			*dest = *p.defaultValue
			return nil
		}
		return d.Err("expected a non-empty value")
	}
	return nil
}

func (p *parseStringMap) parseInline(d *caddyfile.Dispenser, dest *map[string]string) error {
	var empty = true
	for d.NextArg() {
		raw := d.Val()
		if p.keyValueSeparator == "" {
			// default separator is "="
			p.keyValueSeparator = "="
		}
		parts := strings.SplitN(raw, p.keyValueSeparator, 2)
		if len(parts) != 2 {
			return d.Err("expected a key-value separator")
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if _, ok := (*dest)[key]; ok {
			return d.Err("duplicate key")
		}
		if *dest == nil {
			*dest = make(map[string]string)
		}
		(*dest)[key] = value
		empty = false
	}
	if empty && !p.allowEmpty {
		if p.defaultValue != nil {
			*dest = *p.defaultValue
			return nil
		}
		return d.Err("expected a non-empty value")
	}
	return nil
}

var (
	_ allowEmpty     = (*parseStringMap)(nil)
	_ allowDefault   = (*parseStringMap)(nil)
	_ allowInline    = (*parseStringMap)(nil)
	_ allowSeparator = (*parseStringMap)(nil)
)
