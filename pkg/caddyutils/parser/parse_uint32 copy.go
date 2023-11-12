// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseUint64(d *caddyfile.Dispenser, dest *uint64, opts ...Option) error {
	p := &parseUint64{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseUint64 struct {
	allowEmpty   bool
	defaultValue *uint64
	inplace      bool
}

func (p *parseUint64) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseUint64) SetDefaultValue(value interface{}) error {
	v, ok := value.(uint64)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseUint64) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseUint64) parse(d *caddyfile.Dispenser, dest *uint64) error {
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
	*dest = uint64(val)
	return nil
}

func (p *parseUint64) parseInplace(d *caddyfile.Dispenser, dest *uint64) error {
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
	*dest = uint64(val)
	return nil
}
