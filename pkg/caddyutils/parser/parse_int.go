// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseInt(d *caddyfile.Dispenser, dest *int, opts ...Option) error {
	p := &parseInt{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseInt struct {
	allowEmpty   bool
	defaultValue *int
	inplace      bool
}

func (p *parseInt) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseInt) SetDefaultValue(value interface{}) error {
	v, ok := value.(int)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseInt) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseInt) parse(d *caddyfile.Dispenser, dest *int) error {
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
	*dest = val
	return nil
}

func (p *parseInt) parseInplace(d *caddyfile.Dispenser, dest *int) error {
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
	*dest = val
	return nil
}
