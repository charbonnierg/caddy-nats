package caddyutils

import (
	"errors"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/parseutils"
)

func ParseBool(d *caddyfile.Dispenser, dest *bool) error {
	if !d.NextArg() {
		*dest = true
		return nil
	}
	raw := d.Val()
	switch raw {
	case "1", "true", "on", "yes":
		*dest = true
		return nil
	case "0", "false", "off", "no":
		*dest = false
		return nil
	default:
		val, err := strconv.ParseBool(raw)
		if err != nil {
			return d.Errf("invalid boolean value: %s", raw)
		}
		*dest = val
		return nil
	}
}

func ParseBoolReverse(d *caddyfile.Dispenser, dest *bool) error {
	var flag bool
	err := ParseBool(d, &flag)
	if err != nil {
		return err
	}
	*dest = !flag
	return nil
}

func ParseStringArray(d *caddyfile.Dispenser, dest *[]string, allowEmpty bool) error {
	vals := []string{}
	for d.NextArg() {
		if val := d.Val(); val != "" {
			vals = append(vals, val)
		}
	}
	if len(vals) == 0 && !allowEmpty {
		return d.Err("expected at least one string value")
	}
	if dest == nil {
		*dest = []string{}
	}
	*dest = append(*dest, vals...)
	return nil
}

func ExpectString(d *caddyfile.Dispenser, expected string) error {
	if !d.NextArg() {
		return d.Errf("expected %s", expected)
	}
	if d.Val() != expected {
		return d.Errf("expected %s", expected)
	}
	return nil
}

func ParseString(d *caddyfile.Dispenser, dest *string) error {
	if !d.NextArg() {
		return d.Err("expected a string value")
	}
	*dest = d.Val()
	return nil
}

func ParsePermissions(d *caddyfile.Dispenser, dest *fs.FileMode) error {
	if !d.NextArg() {
		return d.Err("expected permission value")
	}
	val, err := strconv.ParseUint(d.Val(), 8, 32)
	if err != nil {
		return d.Errf("invalid permissions value: %s", d.Val())
	}
	*dest = fs.FileMode(val)
	return nil
}

func ParseInt(d *caddyfile.Dispenser, dest *int) error {
	if !d.NextArg() {
		return d.Err("expected a integer value")
	}
	val, err := strconv.Atoi(d.Val())
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = val
	return nil
}

func ParseIntArray(d *caddyfile.Dispenser, dest *[]int) error {
	values := []string{}
	if err := ParseStringArray(d, &values, false); err != nil {
		return err
	}
	ports := make([]int, len(values))
	for idx, value := range values {
		port, err := parseutils.ParsePort(value)
		if err != nil {
			return err
		}
		ports[idx] = port
	}
	*dest = append(*dest, ports...)
	return nil
}

func ParseUInt8(d *caddyfile.Dispenser, dest *uint8) error {
	var value int
	if err := ParseInt(d, &value); err != nil {
		return err
	}
	uint8value, err := parseutils.UInt8(value)
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = uint8value
	return nil
}

func ParseInt64(d *caddyfile.Dispenser, dest *int64) error {
	if !d.NextArg() {
		return d.Err("expected a integer value")
	}
	val, err := strconv.ParseInt(d.Val(), 10, 64)
	if err != nil {
		return d.Errf("invalid integer value: %s", d.Val())
	}
	*dest = val
	return nil
}

func ParseByteSize(d *caddyfile.Dispenser, dest *int) error {
	if !d.NextArg() {
		return d.Err("expected a byte size value")
	}
	val, err := parseutils.ParseBytes(d.Val())
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	*dest = val
	return nil
}

func ParseByteSizeI32(d *caddyfile.Dispenser, dest *int32) error {
	if !d.NextArg() {
		return d.Err("expected a byte size value")
	}
	val, err := parseutils.ParseBytes(d.Val())
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	i32val, err := parseutils.Int32(val)
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	*dest = i32val
	return nil
}

func ParseByteSizeI64(d *caddyfile.Dispenser, dest *int64) error {
	if !d.NextArg() {
		return d.Err("expected a byte size value")
	}
	val, err := parseutils.ParseBytes(d.Val())
	if err != nil {
		return d.Errf("invalid byte size value: %s", d.Val())
	}
	*dest = int64(val)
	return nil
}

func ParsePort(d *caddyfile.Dispenser, dest *int) error {
	if !d.NextArg() {
		return d.Err("expected a port value")
	}
	val, err := parseutils.ParsePort(d.Val())
	if err != nil {
		return d.Errf("invalid port value: %s", d.Val())
	}
	*dest = val
	return nil
}

func ParseDuration(d *caddyfile.Dispenser, dest *time.Duration) error {
	if !d.NextArg() {
		return d.Err("expected a duration value")
	}
	val, err := caddy.ParseDuration(d.Val())
	if err != nil {
		return d.Errf("invalid duration value: %s", d.Val())
	}
	*dest = val
	return nil
}

func ParseSecondsDuration(d *caddyfile.Dispenser, dest *int) error {
	var duration time.Duration
	if err := ParseDuration(d, &duration); err != nil {
		return err
	}
	seconds := int(duration.Seconds())
	if duration > 0 && seconds == 0 {
		seconds = 1
	}
	return nil
}

func ParseKeyValuePairs(d *caddyfile.Dispenser, dest *map[string]string, sep string) error {
	allValues := d.RemainingArgs()
	if len(allValues) == 0 {
		return errors.New("expected at least one key value pair")
	}
	if dest == nil {
		*dest = map[string]string{}
	}
	for _, tag := range allValues {
		if len(tag) == 0 {
			return d.Err("empty tag value")
		}
		keyvalue := strings.Split(tag, sep)
		if len(keyvalue) != 2 {
			return d.Err("invalid tag value")
		}
		key := strings.TrimSpace(keyvalue[0])
		value := strings.TrimSpace(keyvalue[1])
		if len(key) == 0 || len(value) == 0 {
			return d.Err("empty tag key or value")
		}
		if _, ok := (*dest)[key]; ok {
			return d.Err("duplicate tag key")
		}
		(*dest)[key] = value
	}
	return nil
}
