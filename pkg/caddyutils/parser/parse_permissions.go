package parser

import (
	"io/fs"
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParsePermissions(d *caddyfile.Dispenser, dest *fs.FileMode, opts ...Option) error {
	p := &parsePermission{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)
}

type parsePermission struct {
	inplace      bool
	defaultValue *fs.FileMode
}

func (p *parsePermission) SetInplace(value bool) {
	p.inplace = value
}

func (p *parsePermission) SetDefaultValue(value interface{}) error {
	v, ok := value.(fs.FileMode)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parsePermission) parse(d *caddyfile.Dispenser, dest *fs.FileMode) error {
	if !p.inplace {
		if !d.NextArg() {
			if p.defaultValue == nil {
				return d.Err("expected permission value")
			}
			*dest = *p.defaultValue
			return nil
		}
	}
	value := d.Val()
	if value == "" {
		if p.defaultValue == nil {
			return d.Err("expected permission value")
		}
		*dest = *p.defaultValue
		return nil
	}
	val, err := strconv.ParseUint(d.Val(), 8, 32)
	if err != nil {
		return d.Errf("invalid permissions value: %s", d.Val())
	}
	*dest = fs.FileMode(val)
	return nil
}
