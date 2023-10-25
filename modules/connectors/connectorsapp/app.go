package connectorsapp

import (
	"encoding/json"
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/connectors"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(App{})
}

type App struct {
	ctx        caddy.Context
	logger     *zap.Logger
	inputs     []connectors.InputConnector
	outputs    []connectors.OutputConnector
	InputsRaw  []json.RawMessage `json:"inputs,omitempty" caddy:"namespace=connectors.input inline_key=source"`
	OutputsRaw []json.RawMessage `json:"outputs,omitempty" caddy:"namespace=connectors.output inline_key=destination"`
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "connectors",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()
	a.inputs = []connectors.InputConnector{}
	a.outputs = []connectors.OutputConnector{}
	if err := a.provisionInputs(); err != nil {
		return err
	}
	if err := a.provisionOutputs(); err != nil {
		return err
	}
	return nil
}

func (a *App) provisionInputs() error {
	unm, err := a.ctx.LoadModule(a, "InputsRaw")
	if err != nil {
		return err
	}
	for _, raw := range unm.([]interface{}) {
		input, ok := raw.(connectors.InputConnector)
		if !ok {
			return errors.New("input is not a connectors.InputConnector")
		}
		if err := input.Provision(a); err != nil {
			return err
		}
		a.inputs = append(a.inputs, input)
	}
	return nil
}

func (a *App) provisionOutputs() error {
	unm, err := a.ctx.LoadModule(a, "OutputsRaw")
	if err != nil {
		return err
	}
	for _, raw := range unm.([]interface{}) {
		input, ok := raw.(connectors.OutputConnector)
		if !ok {
			return errors.New("ouput is not a connectors.OuputConnector")
		}
		if err := input.Provision(a); err != nil {
			return err
		}
		a.outputs = append(a.outputs, input)
	}
	return nil
}

func (a *App) Context() caddy.Context { return a.ctx }
func (a *App) Start() error {
	for _, input := range a.inputs {
		if err := input.Start(); err != nil {
			return err
		}
	}
	// for _, output := range a.outputs {
	// 	if err := output.Start(); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
func (a *App) Stop() error { return nil }
