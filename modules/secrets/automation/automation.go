// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Automation{})
}

type TriggerKey struct{}

// Automation is a secret automation.
type Automation struct {
	SourcesRaw  []string          `json:"sources,omitempty"`
	TemplateRaw string            `json:"template,omitempty"`
	TriggerRaw  json.RawMessage   `json:"trigger,omitempty" caddy:"namespace=secrets.triggers inline_key=type"`
	HandlersRaw []json.RawMessage `json:"handlers,omitempty" caddy:"namespace=secrets.handlers inline_key=type"`

	app      secrets.App
	ctx      caddy.Context
	cancel   func()
	done     chan struct{}
	logger   *zap.Logger
	trigger  secrets.Trigger
	sources  secrets.Sources
	template *DefaultTemplate
	handlers []secrets.Handler
}

func (Automation) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.automation",
		New: func() caddy.Module { return new(Automation) },
	}
}

// Provision prepares the automation for use.
func (a *Automation) Provision(app secrets.App) error {
	a.app = app
	a.ctx = app.Context()
	a.logger = a.ctx.Logger().Named("automation")
	// Parse raw fields
	if err := a.loadSourcesRaw(); err != nil {
		return err
	}
	if err := a.loadTemplateRaw(); err != nil {
		return err
	}
	if err := a.loadHandlersRaw(); err != nil {
		return err
	}
	if err := a.loadTriggerRaw(); err != nil {
		return err
	}
	return nil
}

// Start starts the automation in background.
func (a *Automation) Start() error {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.done = done
	go a.loop(ctx, done)
	return nil
}

// Stop cancels the automation and waits for it to stop.
func (a *Automation) Stop() error {
	a.cancel()
	<-a.done
	return nil
}

// loop is the main loop of the automation.
func (a *Automation) loop(ctx context.Context, done chan struct{}) {
	channel := a.trigger.Subscribe(ctx)
	for {
		select {
		case <-ctx.Done():
			done <- struct{}{}
			return
		case <-channel:
			a.logger.Info("running secret automation")
			// Fetch secrets
			items, err := a.fetchSecrets()
			if err != nil {
				a.logger.Error("failed to get all secrets", zap.Error(err))
				continue
			}
			// Transform secrets into a string
			value, err := a.formatSecrets(items)
			if err != nil {
				a.logger.Error("failed to format secrets", zap.Error(err))
				continue
			}
			// Call handlers
			err = a.handleSecrets(value)
			if err != nil {
				a.logger.Error("failed to handle secrets", zap.Error(err))
				continue
			}
		}
	}
}

// loadTriggerRaw loads the trigger from the raw configuration.
func (a *Automation) loadTriggerRaw() error {
	unm, err := a.ctx.LoadModule(a, "TriggerRaw")
	if err != nil {
		return err
	}
	trigger, ok := unm.(secrets.Trigger)
	if !ok {
		return fmt.Errorf("trigger is not a Trigger: %T", unm)
	}
	a.trigger = trigger
	return nil
}

// loadSourcesRaw loads the sources from the raw configuration.
func (a *Automation) loadSourcesRaw() error {
	for _, secretRaw := range a.SourcesRaw {
		source, err := a.app.GetSource(secretRaw)
		if err != nil {
			return err
		}
		a.sources = append(a.sources, source)
	}
	return nil
}

// loadTemplateRaw loads the template from the raw configuration.
func (a *Automation) loadTemplateRaw() error {
	if a.TemplateRaw != "" {
		a.template = &DefaultTemplate{
			TemplateBody: a.TemplateRaw,
		}
		err := a.template.Provision(a.app, a)
		if err != nil {
			return err
		}
	}
	return nil
}

// loadHandlersRaw loads the handlers from the raw configuration.
func (a *Automation) loadHandlersRaw() error {
	unm, err := a.ctx.LoadModule(a, "HandlersRaw")
	if err != nil {
		return err
	}
	for _, handlerRaw := range unm.([]interface{}) {
		handler, ok := handlerRaw.(secrets.Handler)
		if !ok {
			return fmt.Errorf("handler is not a Handler: %T", handlerRaw)
		}
		err = handler.Provision(a.app, a)
		if err != nil {
			return err
		}
		a.handlers = append(a.handlers, handler)
	}
	return nil
}

// fetchSecrets fetches all secrets from the sources.
func (a *Automation) fetchSecrets() (secrets.Secrets, error) {
	values := secrets.Secrets{}
	for _, src := range a.sources {
		val, err := src.Get()
		if err != nil {
			a.logger.Error("failed to get secret", zap.String("key", src.Key), zap.String("store", src.StoreName), zap.Error(err))
			return nil, err
		}
		values = append(values, &secrets.Secret{Source: src, Value: val})
	}
	return values, nil
}

// formatSecrets transforms the secrets into a string.
func (a *Automation) formatSecrets(items secrets.Secrets) (string, error) {
	if a.template != nil {
		// If there is a template, we render it
		var err error
		value, err := a.template.Render(items)
		if err != nil {
			return "", err
		}
		return value, nil
	}
	// If there is no template, we just join the values with a newline
	lines := []string{}
	for _, item := range items {
		lines = append(lines, item.Value)
	}
	return strings.Join(lines, "\n"), nil
}

// handleSecrets calls all handlers with the given value.
func (a *Automation) handleSecrets(value string) error {
	for _, handler := range a.handlers {
		_, err := handler.Handle(value)
		if err != nil {
			return err
		}
	}
	return nil
}
