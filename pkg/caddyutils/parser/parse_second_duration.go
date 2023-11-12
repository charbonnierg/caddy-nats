// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseSecondDuration(d *caddyfile.Dispenser, dest *int, opts ...Option) error {
	p := new(parseSecondDuration)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseSecondDuration struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *int
}

func (p *parseSecondDuration) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseSecondDuration) SetDefaultValue(value interface{}) error {
	v, ok := value.(int)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseSecondDuration) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseSecondDuration) parse(d *caddyfile.Dispenser, dest *int) error {
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
	raw := d.Val()
	var val int
	var err error
	val, err = strconv.Atoi(raw)
	if err != nil {
		duration, err := caddy.ParseDuration(raw)
		if err != nil {
			return d.Errf("invalid duration value: %s", raw)
		}
		val = int(duration.Seconds())
		if duration > 0 && val == 0 {
			val = 1
		}
	}
	*dest = val
	return nil
}

func (p *parseSecondDuration) parseInplace(d *caddyfile.Dispenser, dest *int) error {
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
	raw := d.Val()
	var val int
	var err error
	val, err = strconv.Atoi(raw)
	if err != nil {
		duration, err := caddy.ParseDuration(raw)
		if err != nil {
			return d.Errf("invalid duration value: %s", raw)
		}
		val = int(duration.Seconds())
		if duration > 0 && val == 0 {
			val = 1
		}
	}
	*dest = val
	return nil
}
