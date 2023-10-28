// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package automation

import (
	"fmt"
	"os"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
)

// DefaultTemplate is a secret template.
type DefaultTemplate struct {
	template string
	// This is the template to use for the secret
	TemplateFile string `json:"template_file,omitempty"`
	TemplateBody string `json:"template_body,omitempty"`
}

// CaddyModule returns the caddy module information.
func (t *DefaultTemplate) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.template.default",
		New: func() caddy.Module { return new(DefaultTemplate) },
	}
}

// Provision prepares the template for use.
func (t *DefaultTemplate) Provision(_ secrets.App, __ secrets.Automation) error {
	switch {
	// Both template_file and template are set
	case t.TemplateFile != "" && t.TemplateBody != "":
		return fmt.Errorf("template_body and template_file are mutually exclusive")
	// template_file is set
	case t.TemplateFile != "":
		return t.loadTemplateFile()
	// template is set
	case t.TemplateBody != "":
		t.template = t.TemplateBody
		return nil
	// template_file and template are not set
	default:
		return fmt.Errorf("template_file or template_body must be set")
	}
}

// Render renders the template with the given values.
func (t *DefaultTemplate) Render(input secrets.Secrets) (string, error) {
	res, err := t.newReplacer(input).ReplaceOrErr(t.template, true, true)
	if err != nil {
		return "", err
	}
	return res, nil
}

// Private method to load the template from a file.
func (t *DefaultTemplate) loadTemplateFile() error {
	content, err := os.ReadFile(t.TemplateFile)
	if err != nil {
		return err
	}
	t.template = string(content)
	return nil
}

// Private method to create a replacer according to some input
func (t *DefaultTemplate) newReplacer(input secrets.Secrets) *caddy.Replacer {
	repl := caddy.NewReplacer()
	for _, secret := range input {
		repl.Set(secret.Source.String(), secret.Value)
	}
	now := time.Now()
	repl.Map(func(key string) (any, bool) {
		switch key {
		case "now.Unix()":
			return now.Unix(), true
		case "now.UnixNano()":
			return now.UnixNano(), true
		case "now.RFC1123()":
			return now.Format(time.RFC1123), true
		case "now.RFC3339()":
			return now.Format(time.RFC3339), true
		case "now.RFC339Nano()":
			return now.Format(time.RFC3339Nano), true
		case "now.RFC3339Nano()":
			return now.Format(time.RFC3339Nano), true
		case "now.RFC822()":
			return now.Format(time.RFC822), true
		case "now.RFC822Z()":
			return now.Format(time.RFC822Z), true
		case "now.RFC850()":
			return now.Format(time.RFC850), true
		}
		return nil, false
	})
	return repl
}

var (
	_ secrets.Template = (*DefaultTemplate)(nil)
)
