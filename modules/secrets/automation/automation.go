// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package automation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

// Handler is a secret automation handler.
// It is used to handle the secret value when it is fetched.
type Handler interface {
	caddy.Module
	Provision(app secrets.SecretApp, automation *Automation) error
	Handle(value string) (string, error)
}

// Automation is a secret automation.
// It periodically fetches secrets from sources and executes handlers with the secrets.
// Optionally, a template can be used to transform the secret values into a custom string.
// By default, the secret values are joined with a newline.
type Automation struct {
	app      secrets.SecretApp
	ctx      caddy.Context
	logger   *zap.Logger
	sources  []*secrets.Source
	template *Template
	handlers []Handler

	Interval    time.Duration     `json:"interval,omitempty"`
	SourcesRaw  []string          `json:"sources,omitempty"`
	TemplateRaw string            `json:"template,omitempty"`
	HandlersRaw []json.RawMessage `json:"handlers,omitempty" caddy:"namespace=secrets.handlers inline_key=type"`
}

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

func (a *Automation) loadTemplateRaw() error {
	if a.TemplateRaw != "" {
		a.template = &Template{
			TemplateRaw: a.TemplateRaw,
		}
		err := a.template.Provision(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Automation) loadHandlersRaw() error {
	unm, err := a.ctx.LoadModule(a, "HandlersRaw")
	if err != nil {
		return err
	}
	for _, handlerRaw := range unm.([]interface{}) {
		handler, ok := handlerRaw.(Handler)
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

// Provision prepares the automation for use.
func (a *Automation) Provision(app secrets.SecretApp) error {
	a.app = app
	a.ctx = app.Context()
	a.logger = a.ctx.Logger().Named("secrets.automation")
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
	return nil
}

// Run starts the automation, and keeps running until the context is cancelled.
// There is no need to stop the automation, it will stop automatically when the context is cancelled.
func (a *Automation) Run() error {
	ticker := time.NewTicker(a.Interval)
	for {
		select {
		case <-a.ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			a.logger.Info("running secret automation")
			willRetry := false
			// Fetch secrets
			values := map[string]string{}
			for _, secret := range a.sources {
				val, err := secret.Store.Get(secret.Key)
				if err != nil {
					a.logger.Error("failed to get secret", zap.String("key", secret.Key), zap.Error(err))
					willRetry = true
					break
				}
				values[secret.Key] = val
			}
			if willRetry {
				a.logger.Error("failed to get all secrets", zap.Duration("next_retry", a.Interval))
				continue
			}
			// Transform secrets into a string
			var value string
			switch a.template {
			case nil:
				// If there is no template, we just join the values with a newline
				lines := []string{}
				for k, v := range values {
					lines = append(lines, fmt.Sprintf("%s=%s", k, v))
				}
				value = strings.Join(lines, "\n")
			default:
				// If there is a template, we render it
				var err error
				value, err = a.template.Render(values)
				if err != nil {
					a.logger.Error("failed to render template", zap.Error(err), zap.Duration("next_retry", a.Interval))
					continue
				}
			}
			// Call handlers
			var err error
			for _, handler := range a.handlers {
				value, err = handler.Handle(value)
				if err != nil {
					a.logger.Error("failed to handle secret", zap.Error(err), zap.Duration("next_retry", a.Interval))
					continue
				}
			}
		}
	}
}
