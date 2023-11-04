package app

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func parseGlobalOption(d *caddyfile.Dispenser, existingVal interface{}) (interface{}, error) {
	a := new(App)
	if existingVal != nil {
		var ok bool
		caddyFileApp, ok := existingVal.(httpcaddyfile.App)
		if !ok {
			return nil, d.Errf("existing secrets app of unexpected type: %T", existingVal)
		}
		err := json.Unmarshal(caddyFileApp.Value, a)
		if err != nil {
			return nil, err
		}
	}
	err := a.UnmarshalCaddyfile(d)
	return httpcaddyfile.App{
		Name:  "python",
		Value: caddyconfig.JSON(a, nil),
	}, err
}

func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d, parser.Match("python")); err != nil {
		return err
	}
	a.Processes = fnutils.DefaultIfEmpty(a.Processes, []*PythonProcess{})
	defaultVenv := ""
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "virtualenv":
			if err := parser.ParseString(d, &defaultVenv); err != nil {
				return err
			}
		case "process", "app":
			process := &PythonProcess{VirtualEnv: defaultVenv}
			if err := parser.ParseString(d, &process.Name); err != nil {
				return err
			}
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "entrypoint":
					process.Command = "python3"
					if err := parser.ParseStringArray(d, &process.Args, parser.AllowEmpty()); err != nil {
						return err
					}
				case "command":
					if err := parser.ParseString(d, &process.Command); err != nil {
						return err
					}
					if err := parser.ParseStringArray(d, &process.Args, parser.AllowEmpty()); err != nil {
						return err
					}
				case "virtualenv":
					if err := parser.ParseString(d, &process.VirtualEnv); err != nil {
						return err
					}
				case "working_dir":
					if err := parser.ParseString(d, &process.WorkingDir); err != nil {
						return err
					}
				case "environment":
					if err := parser.ParseStringMap(d, &process.Environment); err != nil {
						return err
					}
				case "forward_stderr":
					if err := parser.ParseBool(d, &process.ForwardStderr); err != nil {
						return err
					}
				case "forward_stdout":
					if err := parser.ParseBool(d, &process.ForwardStdout); err != nil {
						return err
					}
				default:
					return d.Errf("unknown python process option '%s'", d.Val())
				}
			}
			a.Processes = append(a.Processes, process)
		default:
			return d.Errf("unknown python directive '%s'", d.Val())
		}
	}
	return nil
}
