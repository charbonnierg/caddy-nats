// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func ParseInt32ByteSize(d *caddyfile.Dispenser, dest *int32, opts ...Option) error {
	p := new(parseInt32ByteSize)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseInt32ByteSize struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *int32
}

func (p *parseInt32ByteSize) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseInt32ByteSize) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseInt32ByteSize) SetDefaultValue(value interface{}) error {
	v, ok := value.(int32)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseInt32ByteSize) parse(d *caddyfile.Dispenser, dest *int32) error {
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
	*dest = int32(val)
	return nil
}

func (p *parseInt32ByteSize) parseInplace(d *caddyfile.Dispenser, dest *int32) error {
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
	*dest = int32(val)
	return nil
}
