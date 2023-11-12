// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ParseTime(d *caddyfile.Dispenser, dest *time.Time, opts ...Option) error {
	p := &parseTime{}
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return err
		}
	}
	return p.parse(d, dest)

}

type parseTime struct{}

func (p *parseTime) parse(d *caddyfile.Dispenser, dest *time.Time) error {
	return nil
}
