package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

type parseNetworkPortRange struct {
	allowEmpty   bool
	defaultValue *[]int
	inplace      bool
}

func (p *parseNetworkPortRange) SetAllowEmpty(allowEmpty bool) {
	p.allowEmpty = allowEmpty
}

func (p *parseNetworkPortRange) SetDefaultValue(value interface{}) error {
	v, ok := value.([]int)
	if !ok {
		return ErrInvalidDefaultType
	}
	p.defaultValue = &v
	return nil
}

func (p *parseNetworkPortRange) SetInplace(value bool) {
	p.inplace = value
}

func (p *parseNetworkPortRange) parse(d *caddyfile.Dispenser, dest *[]int) error {
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
	ports, err := parsePortRange(d.Val())
	if err != nil {
		return d.Errf("invalid port range: %s", err)
	}
	if *dest == nil {
		*dest = ports
	} else {
		*dest = append(*dest, ports...)
	}
	return nil
}

func (p *parseNetworkPortRange) parseInplace(d *caddyfile.Dispenser, dest *[]int) error {
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
	ports, err := parsePortRange(d.Val())
	if err != nil {
		return d.Errf("invalid port range: %s", err)
	}
	if *dest == nil {
		*dest = ports
	} else {
		*dest = append(*dest, ports...)
	}
	return nil
}

func parsePortRange(s string) ([]int, error) {
	var allPorts []int
	ranges, err := splitPorts(s)
	if err != nil {
		return nil, err
	}
	for _, portRange := range ranges {
		ports, err := parseRange(portRange)
		if err != nil {
			return nil, err
		}
		allPorts = append(allPorts, ports...)
	}
	return allPorts, nil
}

func splitPorts(s string) ([][]string, error) {
	// Split by comma first, then by dash
	ranges := strings.Split(s, ",")
	subranges := [][]string{}
	for _, r := range ranges {
		subrange := strings.Split(r, "-")
		subranges = append(subranges, subrange)
	}
	return subranges, nil
}

func parseRange(s []string) ([]int, error) {
	var ports []int
	switch len(s) {
	// This is a single port, e.g. 80
	case 1:
		p, err := parsePortValue(s[0])
		if err != nil {
			return nil, fmt.Errorf("invalid port range: %w", err)
		}
		ports = append(ports, p)
	// This is a range of ports, e.g. 80-90
	case 2:
		start, err := parsePortValue(s[0])
		if err != nil {
			return nil, fmt.Errorf("invalid port range: %w", err)
		}
		end, err := parsePortValue(s[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port range: %w", err)
		}
		if start > end {
			return nil, errors.New("invalid port range")
		}
		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
	// This is an invalid range, e.g. 80-90-100
	default:
		return nil, errors.New("invalid port range")
	}
	return ports, nil
}

func parsePortValue(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid port value: %w", err)
	}
	if p < 0 || p > 65535 {
		return 0, errors.New("invalid port value")
	}
	return p, nil
}
