package caddyutils

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseBool(d *caddyfile.Dispenser) (bool, error) {
	if !d.NextArg() {
		return true, nil
	}
	raw := d.Val()
	switch raw {
	case "1", "true", "on", "yes":
		return true, nil
	case "0", "false", "off", "no":
		return false, nil
	default:
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return false, d.Errf("invalid boolean value: %s", raw)
		}
		return val, nil
	}
}

func ParseStringArray(d *caddyfile.Dispenser) []string {
	vals := []string{}
	for d.NextArg() {
		if val := d.Val(); val != "" {
			vals = append(vals, val)
		}
	}
	return vals
}
