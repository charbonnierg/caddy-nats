package automation

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Template is a secret template.
type Template struct {
	template *template.Template
	// This is the template to use for the secret
	TemplateFile string `json:"template_file,omitempty"`
	TemplateRaw  string `json:"template,omitempty"`
}

// Private method to load the template from a file.
func (t *Template) loadTemplateFile() error {
	if t.TemplateRaw != "" {
		return fmt.Errorf("template_file and template are mutually exclusive")
	}
	tmpl, err := template.ParseFiles(t.TemplateFile)
	if err != nil {
		return err
	}
	t.template = tmpl
	return nil
}

// Private method to load the template from a string.
func (t *Template) loadTemplateRaw() error {
	tmpl, err := template.New("secret").Parse(t.TemplateRaw)
	if err != nil {
		return err
	}
	t.template = tmpl
	return nil
}

// Provision prepares the template for use.
func (t *Template) Provision(automation *Automation) error {
	switch {
	// Both template_file and template are set
	case t.TemplateFile != "" && t.TemplateRaw != "":
		return fmt.Errorf("template_file and template are mutually exclusive")
	// template_file is set
	case t.TemplateFile != "":
		return t.loadTemplateFile()
	// template is set
	case t.TemplateRaw != "":
		return t.loadTemplateRaw()
	// template_file and template are not set
	default:
		return fmt.Errorf("template_file or template must be set")
	}
}

// Render renders the template with the given values.
func (t *Template) Render(value map[string]string) (string, error) {
	data := map[string]any{}
	data["values"] = value
	for k, v := range value {
		_k := strings.Split(k, "@")
		data[strings.ReplaceAll(_k[0], "-", "_")] = v
	}
	buf := bytes.NewBuffer(nil)
	err := t.template.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
