// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseInt64(d *caddyfile.Dispenser, dest *int64, opts ...Option) error {
	p := &parseInt64{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseInt64 struct {
	allowEmpty   bool
	defaultValue *int64
	inplace      bool
}

func (p *parseInt64) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseInt64) SetDefaultValue(value interface{}) error {
	v, ok := value.(int64)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseInt64) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseInt64) parse(d *caddyfile.Dispenser, dest *int64) error {
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
	val, err := strconv.ParseInt(d.Val(), 10, 64)
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = val
	return nil
}

func (p *parseInt64) parseInplace(d *caddyfile.Dispenser, dest *int64) error {
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
	val, err := strconv.ParseInt(d.Val(), 10, 64)
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = val
	return nil
}
