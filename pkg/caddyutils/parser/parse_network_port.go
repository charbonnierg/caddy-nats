package parser

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseNetworkPort(d *caddyfile.Dispenser, dest *int, opts ...Option) error {
	p := &parseNetworkPort{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parseNetworkPort struct {
	allowEmpty   bool
	defaultValue *int
	inplace      bool
}

func (p *parseNetworkPort) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseNetworkPort) SetDefaultValue(value interface{}) error {
	v, ok := value.(int)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseNetworkPort) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseNetworkPort) parse(d *caddyfile.Dispenser, dest *int) error {
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
	port, err := parsePortValue(d.Val())
	if err != nil {
		return d.Errf("invalid port value: %w", err)
	}
	*dest = port
	return nil
}

func (p *parseNetworkPort) parseInplace(d *caddyfile.Dispenser, dest *int) error {
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
	port, err := parsePortValue(d.Val())
	if err != nil {
		return d.Errf("invalid port value: %w", err)
	}
	*dest = port
	return nil
}

func ParseNetworkPortRange(d *caddyfile.Dispenser, dest *[]int, opts ...Option) error {
	p := &parseNetworkPortRange{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}
