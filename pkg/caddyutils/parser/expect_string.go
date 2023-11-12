// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"strings"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/quara-dev/beyond/pkg/datatypes"
)

// ExpectString loads the next argument from the dispenser and
// verify that it is not empty by default.
// When Match option is provided, an error is raised if argument
// is not equal to one of the strings to match:
//
//	err := parser.ExpectString(d, parser.Match("value1", "value2"))
//
// will return an error if next argument is not "value1" or "value2".
//
// When the Inplace option is used, next argument is not loaded, and
// the value currently loaded is used instead:
//
//	err := parser.ExpectString(d, parser.Match("somevalue"), parser.Inplace())
func ExpectString(d *caddyfile.Dispenser, opts ...Option) error {
	p := &expectString{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d)
}

type expectString struct {
	inplace bool
	match   datatypes.StringSet
}

func (p *expectString) expected() string {
	if len(p.match) == 0 {
		return "a string"
	}
	return "one of: " + strings.Join(p.match.Slice(), ", ")
}

func (p *expectString) SetInplace(value bool) {
	p.inplace = value
}

func (p *expectString) AddMatch(match string, other ...string) error {
	if p.match == nil {
		p.match = make(datatypes.StringSet)
	}
	p.match.Add(match, other...)
	return nil
}

func (p *expectString) parse(d *caddyfile.Dispenser) error {
	if !p.inplace && !d.NextArg() {
		return d.Errf("expected %s", p.expected())
	}
	if len(p.match) == 0 {
		return nil
	}
	val := d.Val()
	if p.match.Contains(val) {
		return nil
	}
	return d.Errf("expected %s, but got %s", p.expected(), val)
}

var (
	_ allowMatch   = (*expectString)(nil)
	_ allowInplace = (*expectString)(nil)
)
