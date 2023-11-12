// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseUint(d *caddyfile.Dispenser, dest *uint, opts ...Option) error {
	p := &parseUInt{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseUInt struct {
	allowEmpty   bool
	defaultValue *uint
	inplace      bool
}

func (p *parseUInt) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseUInt) SetDefaultValue(value interface{}) error {
	v, ok := value.(uint)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseUInt) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseUInt) parse(d *caddyfile.Dispenser, dest *uint) error {
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
	*dest = uint(val)
	return nil
}

func (p *parseUInt) parseInplace(d *caddyfile.Dispenser, dest *uint) error {
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
	*dest = uint(val)
	return nil
}
