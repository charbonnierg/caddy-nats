// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseUint8(d *caddyfile.Dispenser, dest *uint8, opts ...Option) error {
	p := &parseUInt8{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseUInt8 struct {
	allowEmpty   bool
	defaultValue *uint8
	inplace      bool
}

func (p *parseUInt8) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseUInt8) SetDefaultValue(value interface{}) error {
	v, ok := value.(uint8)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseUInt8) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseUInt8) parse(d *caddyfile.Dispenser, dest *uint8) error {
	if p.inplace {
		return p.parseInplace(d, dest)
	}
	loaded := d.NextArg()

	if !loaded {
		if p.defaultValue != nil {
			*dest = *p.defaultValue
			return nil
		}
		if p.allowEmpty {
			return nil
		}
		return d.Err("empty value")
	}
	val, err := strconv.Atoi(d.Val())
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	if val < 0 || val > 255 {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = uint8(val)
	return nil
}

func (p *parseUInt8) parseInplace(d *caddyfile.Dispenser, dest *uint8) error {
	if d.Val() == "" {
		if p.defaultValue != nil {
			*dest = *p.defaultValue
			return nil
		}
		if p.allowEmpty {
			return nil
		}
		return d.Err("empty value")
	}
	val, err := strconv.Atoi(d.Val())
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	if val < 0 || val > 255 {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = uint8(val)
	return nil
}
