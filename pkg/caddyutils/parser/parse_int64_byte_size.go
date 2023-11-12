// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func ParseInt64ByteSize(d *caddyfile.Dispenser, dest *int64, opts ...Option) error {
	p := new(parseInt64ByteSize)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseInt64ByteSize struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *int64
}

func (p *parseInt64ByteSize) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseInt64ByteSize) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseInt64ByteSize) SetDefaultValue(value interface{}) error {
	v, ok := value.(int64)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseInt64ByteSize) parse(d *caddyfile.Dispenser, dest *int64) error {
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
	*dest = int64(val)
	return nil
}

func (p *parseInt64ByteSize) parseInplace(d *caddyfile.Dispenser, dest *int64) error {
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
	*dest = int64(val)
	return nil
}
