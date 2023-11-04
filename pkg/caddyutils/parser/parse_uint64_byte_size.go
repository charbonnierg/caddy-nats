package parser

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func ParseUint64ByteSize(d *caddyfile.Dispenser, dest *uint64, opts ...Option) error {
	p := new(parseUint64ByteSize)
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseUint64ByteSize struct {
	inplace      bool
	allowEmpty   bool
	defaultValue *uint64
}

func (p *parseUint64ByteSize) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseUint64ByteSize) SetAllowEmpty(value bool) {
	p.allowEmpty = value
}

func (p *parseUint64ByteSize) SetDefaultValue(value interface{}) error {
	v, ok := value.(uint64)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseUint64ByteSize) parse(d *caddyfile.Dispenser, dest *uint64) error {
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
	*dest = uint64(val)
	return nil
}

func (p *parseUint64ByteSize) parseInplace(d *caddyfile.Dispenser, dest *uint64) error {
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
	*dest = uint64(val)
	return nil
}
