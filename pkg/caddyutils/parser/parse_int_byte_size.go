// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func ParseIntByteSize(d *caddyfile.Dispenser, dest *int, opts ...Option) error {
	p := new(parseIntByteSize)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseIntByteSize struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *int
}

func (p *parseIntByteSize) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseIntByteSize) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseIntByteSize) SetDefaultValue(value interface{}) error {
	v, ok := value.(int)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseIntByteSize) parse(d *caddyfile.Dispenser, dest *int) error {
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
	val, err := parseutils.ParseBytes(d.Val())
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	*dest = val
	return nil
}

func (p *parseIntByteSize) parseInplace(d *caddyfile.Dispenser, dest *int) error {
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
	val, err := parseutils.ParseBytes(d.Val())
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	*dest = val
	return nil
}
