// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseUint32(d *caddyfile.Dispenser, dest *uint32, opts ...Option) error {
	p := &parseUInt32{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseUInt32 struct {
	allowEmpty   bool
	defaultValue *uint32
	inplace      bool
}

func (p *parseUInt32) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseUInt32) SetDefaultValue(value interface{}) error {
	v, ok := value.(uint32)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseUInt32) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseUInt32) parse(d *caddyfile.Dispenser, dest *uint32) error {
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
	if val < 0 {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = uint32(val)
	return nil
}

func (p *parseUInt32) parseInplace(d *caddyfile.Dispenser, dest *uint32) error {
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
	if val < 0 {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = uint32(val)
	return nil
}
