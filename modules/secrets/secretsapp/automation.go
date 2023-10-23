package secretsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

type Handler interface {
	caddy.Module
	Provision(automation *Automation) error
	Handle(value string) (string, error)
}

type Secret struct {
	Store secrets.Store
	Key   string
}

type Template struct {
	template *template.Template
	// This is the template to use for the secret
	RawTemplate string `json:"template,omitempty"`
}

func (t *Template) Provision(automation *Automation) error {
	tmpl, err := template.New("secret").Parse(t.RawTemplate)
	if err != nil {
		return err
	}
	t.template = tmpl
	return nil
}

func (t *Template) Render(value map[string]string) (string, error) {
	data := map[string]string{}
	buf := bytes.NewBuffer(nil)
	err := t.template.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type Automation struct {
	ctx      caddy.Context
	logger   *zap.Logger
	secrets  []Secret
	template *Template
	handlers []Handler
	// Public fields
	Interval    time.Duration     `json:"interval,omitempty"`
	SecretsRaw  []string          `json:"sources,omitempty"`
	TemplateRaw string            `json:"template,omitempty"`
	HandlersRaw []json.RawMessage `json:"handlers,omitempty" caddy:"namespace=secrets.handlers inline_key=type"`
}

func (a *Automation) Provision(app *App) error {
	a.ctx = app.Context()
	a.logger = app.Logger().Named("automation")
	// Parse secrets
	for _, secretRaw := range a.SecretsRaw {
		store, key, err := app.getStoreAndKey(secretRaw)
		if err != nil {
			return err
		}
		a.secrets = append(a.secrets, Secret{
			Store: store,
			Key:   key,
		})
	}
	// Parse template
	if a.TemplateRaw != "" {
		a.template = &Template{
			RawTemplate: a.TemplateRaw,
		}
		err := a.template.Provision(a)
		if err != nil {
			return err
		}
	}
	// Load handlers
	unm, err := a.ctx.LoadModule(a, "HandlersRaw")
	if err != nil {
		return err
	}
	for _, handlerRaw := range unm.([]interface{}) {
		handler, ok := handlerRaw.(Handler)
		if !ok {
			return fmt.Errorf("handler is not a Handler: %T", handlerRaw)
		}
		err = handler.Provision(a)
		if err != nil {
			return err
		}
		a.handlers = append(a.handlers, handler)
	}
	return nil
}

// Run starts the automation, and keeps running until the context is cancelled.
// There is no need to stop the automation, it will stop automatically when the context is cancelled.
func (a *Automation) Run() error {
	timer := time.NewTimer(a.Interval)
	for {
		select {
		case <-a.ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
			willRetry := false
			// Fetch secrets
			values := map[string]string{}
			for _, secret := range a.secrets {
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
