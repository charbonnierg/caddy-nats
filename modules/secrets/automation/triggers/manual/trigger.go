// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package manual

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/modules/secrets/automation"
)

func init() {
	caddy.RegisterModule(ManualTrigger{})
}

// ManualTrigger is a trigger that can be manually triggered during tests.
type ManualTrigger struct {
	channel chan context.Context
}

func (ManualTrigger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.triggers.manual",
		New: func() caddy.Module { return new(ManualTrigger) },
	}
}

func (t *ManualTrigger) Provision(_ secrets.App, __ secrets.Automation) error {
	t.channel = make(chan context.Context)
	return nil
}

func (t *ManualTrigger) Subscribe(ctx context.Context) <-chan context.Context {
	t.channel = make(chan context.Context)
	return t.channel
}

func (t *ManualTrigger) Trigger() {
	t.channel <- context.WithValue(context.Background(), automation.TriggerKey{}, t)
}

var (
	_ secrets.Trigger = (*ManualTrigger)(nil)
)
