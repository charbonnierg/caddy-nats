// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseFloat64(d *caddyfile.Dispenser, dest *float64, opts ...Option) error {
	p := &parseFloat64{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseFloat64 struct {
	allowEmpty   bool
	defaultValue *float64
	inplace      bool
}

func (p *parseFloat64) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseFloat64) SetDefaultValue(value interface{}) error {
	v, ok := value.(float64)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseFloat64) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseFloat64) parse(d *caddyfile.Dispenser, dest *float64) error {
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
	val, err := strconv.ParseFloat(d.Val(), 64)
	if err != nil {
		return d.Errf("invalid float value: %s", d.Val())
	}
	*dest = val
	return nil
}

func (p *parseFloat64) parseInplace(d *caddyfile.Dispenser, dest *float64) error {
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
	val, err := strconv.ParseFloat(d.Val(), 64)
	if err != nil {
		return d.Errf("invalid float value: %s", d.Val())
	}
	*dest = val
	return nil
}
