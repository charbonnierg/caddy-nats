package parser

import (
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// ParseDuration loads the next argument and parses a time duration value from the dispenser
// by default.
//
//   - Default(): sets the default value of the boolean if no value is provided
//
//   - AllowEmpty(): allow empty or missing value. By default empty value is considered an error.
//
//   - Inplace(): do not load next argument but parse current value instead
func ParseDuration(d *caddyfile.Dispenser, dest *time.Duration, opts ...Option) error {
	p := new(parseDuration)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseDuration struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *time.Duration
}

func (p *parseDuration) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseDuration) SetDefaultValue(value interface{}) error {
	v, ok := value.(time.Duration)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseDuration) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseDuration) parse(d *caddyfile.Dispenser, dest *time.Duration) error {
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
	val, err := caddy.ParseDuration(d.Val())
	if err != nil {
		return d.Errf("invalid duration value: %s", d.Val())
	}
	*dest = val
	return nil
}

func (p *parseDuration) parseInplace(d *caddyfile.Dispenser, dest *time.Duration) error {
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
	val, err := caddy.ParseDuration(d.Val())
	if err != nil {
		return d.Errf("invalid float value: %s", d.Val())
	}
	*dest = val
	return nil
}
