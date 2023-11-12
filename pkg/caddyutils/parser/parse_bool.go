// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// ParseBool loads the next argument and parses a boolean value from the dispenser by default
// The following options can be used:
//
//   - Reverse(): reverses the value of the boolean (true becomes false and vice versa)
//
//   - Default(): sets the default value of the boolean if no value is provided
//
//   - ErrorIfEmpty(): disallows empty values. By default empty value is considered true (or false if reversed)
//
//   - Inplace(): do not load next argument but parse current value instead
//
// Example usage:
//
//	type MyModule struct {
//		Enabled bool
//	}
//
//	func (m *MyModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
//		// Skip first token (module name in general)
//		if err := parser.ExpectString(d); err != nil {
//			return err
//		}
//		// Parse block
//		for nesting := d.Nesting(); d.NextBlock(nesting); {
//			switch d.Val() {
//				case "enabled":
//					// Use ParseBool
//					if err := parser.ParseBool(d, &m.Enabled); err != nil {
//						return err
//					}
//				default:
//					return d.Errf("invalid option: %s", d.Val())
//			}
//		}
//		return nil
//	}
//
// Note: When Reverse() is used and a default is provided, the default value is NOT reversed.
func ParseBool(d *caddyfile.Dispenser, dest *bool, opts ...Option) error {
	p := &parseBool{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

// parseBool implementation
type parseBool struct {
	reverse      bool
	defaultValue *bool
	errorIfEmpty bool
}

func (p *parseBool) SetReversed(reverse bool) {
	p.reverse = reverse
}

func (p *parseBool) SetDefaultValue(value any) error {
	v, ok := value.(bool)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseBool) SetAllowEmpty(allowEmpty bool) {
	p.errorIfEmpty = !allowEmpty
}

func (p *parseBool) parse(d *caddyfile.Dispenser, dest *bool) error {
	set := func(v bool) {
		if p.reverse {
			*dest = !v
		} else {
			*dest = v
		}
	}
	if !d.NextArg() {
		if p.errorIfEmpty {
			return d.Err("expected a boolean value")
		}
		if p.defaultValue == nil {
			set(true)
		} else {
			*dest = *p.defaultValue
		}
		return nil
	}
	raw := d.Val()
	switch raw {
	case "1", "true", "on", "yes":
		set(true)
		return nil
	case "0", "false", "off", "no":
		set(false)
		return nil
	default:
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return d.Errf("invalid boolean value: %s", raw)
		}
		set(val)
		return nil
	}
}

var (
	_ allowDefault = (*parseBool)(nil)
	_ allowReverse = (*parseBool)(nil)
	_ allowEmpty   = (*parseBool)(nil)
)
