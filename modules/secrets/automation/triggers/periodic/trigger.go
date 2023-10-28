// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package periodic

import (
	"context"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"github.com/quara-dev/beyond/modules/secrets/automation"
)

func init() {
	caddy.RegisterModule(PeriodicTrigger{})
}

// PeriodicTrigger is a trigger that periodically triggers the automation.
type PeriodicTrigger struct {
	Interval time.Duration `json:"interval,omitempty"`
}

func (PeriodicTrigger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.triggers.periodic",
		New: func() caddy.Module { return new(PeriodicTrigger) },
	}
}

func (t *PeriodicTrigger) Provision(_ secrets.App, __ secrets.Automation) error {
	return nil
}

func (t *PeriodicTrigger) Subscribe(ctx context.Context) <-chan context.Context {
	channel := make(chan context.Context)
	go func() {
		ticker := time.NewTicker(t.Interval)
		unitCtx := context.WithValue(ctx, automation.TriggerKey{}, t)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				close(channel)
				return
			case <-ticker.C:
				channel <- unitCtx
			}
		}
	}()
	return channel
}

var (
	_ secrets.Trigger = (*PeriodicTrigger)(nil)
)
