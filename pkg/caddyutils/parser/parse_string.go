package parser

import (
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/datatypes"
)

func ParseString(d *caddyfile.Dispenser, dest *string, opts ...Option) error {
	p := &parseString{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseString struct {
	inplace      bool
	match        datatypes.StringSet
	allowEmpty   bool
	defaultValue *string
}

func (p *parseString) expected() string {
	return "one of: " + strings.Join(p.match.Slice(), ", ")
}

func (p *parseString) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseString) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseString) SetDefaultValue(value interface{}) error {
	v, ok := value.(string)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseString) AddMatch(value string, values ...string) error {
	if p.match == nil {
		p.match = make(datatypes.StringSet)
	}
	p.match.Add(value, values...)
	return nil
}

func (p *parseString) parse(d *caddyfile.Dispenser, dest *string) error {
	if !p.inplace {
		loaded := d.NextArg()
		if !loaded {
			if p.defaultValue != nil {
				*dest = *p.defaultValue
				return nil
			}
			if p.allowEmpty {
				*dest = ""
				return nil
			}
			return d.Err("expected a string value")
		}
		value := d.Val()
		switch value {
		case "":
			if p.allowEmpty {
				*dest = ""
				return nil
			}
			if p.defaultValue != nil {
				*dest = *p.defaultValue
				return nil
			}
			return d.Err("expected a string value")
		default:
			if len(p.match) > 0 && !p.match.Contains(value) {
				return d.Errf("value must be one of: %s", p.expected())
			}
			*dest = value
			return nil
		}
	}
	value := d.Val()
	switch value {
	case "":
		if p.defaultValue != nil {
			*dest = *p.defaultValue
			return nil
		}
		if p.allowEmpty {
			*dest = ""
			return nil
		}
		return d.Err("expected a string value")
	default:
		if len(p.match) > 0 && !p.match.Contains(value) {
			return d.Errf("value must be one of: %s", p.expected())
		}
		*dest = value
		return nil
	}
}
