// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/secrets"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ExecHandler{})
}

type ExecHandler struct {
	ctx    caddy.Context
	logger *zap.Logger
	// This is the command to execute
	Command string `json:"command,omitempty"`
	// This is the arguments to pass to the command
	Args []string `json:"args,omitempty"`
	// This is the environment variables to forward to the command
	// By default, all environment variables are passed to the command
	// This option allows to specify which environment variables to pass
	// Unknown environment variables are ignored
	// Empty environment variables are forwarded as empty environment variables
	ForwardEnv []string `json:"forward_env,omitempty"`
	// This is the environment variables to set for the command
	// Those environment variables are not kept in the process environment
	Env map[string]string `json:"env,omitempty"`
	// This is the working directory to use when executing the command
	WorkingDir string `json:"working_dir,omitempty"`
}

func (ExecHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets.handlers.exec",
		New: func() caddy.Module { return new(ExecHandler) },
	}
}

func (h *ExecHandler) Provision(app secrets.App, auto secrets.Automation) error {
	h.ctx = app.Context()
	h.logger = h.ctx.Logger().Named("automation.exec_handler")
	if h.WorkingDir != "" {
		_, err := os.Stat(h.WorkingDir)
		if err != nil {
			return fmt.Errorf("failed to stat working directory: %w", err)
		}
	}
	return nil
}

func (h *ExecHandler) Handle(value string) (string, error) {
	cmd, err := h.createCommand(value)
	if err != nil {
		return "", fmt.Errorf("failed to create command: %w", err)
	}
	out, err := cmd.Output()
	if err != nil {
		return string(out), fmt.Errorf("failed to execute command: %w", err)
	}
	result := strings.TrimSpace(string(out))
	return result, nil
}

func (h *ExecHandler) createCommand(value string) (*exec.Cmd, error) {
	repl := caddy.NewReplacer()
	repl.Set("value", value)
	prog := h.Command
	env := h.gatherEnvironmentVariables()
	args, err := h.gatherArguments(repl)
	if err != nil {
		return nil, fmt.Errorf("failed to gather arguments: %w", err)
	}
	cmd := exec.Command(prog, args...)
	cmd.Env = env
	cmd.Dir = h.WorkingDir
	return cmd, nil
}

func (h *ExecHandler) gatherArguments(repl *caddy.Replacer) ([]string, error) {
	args := []string{}
	for _, rawArg := range h.Args {
		arg, err := repl.ReplaceOrErr(rawArg, true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to replace argument: %w", err)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (h *ExecHandler) gatherEnvironmentVariables() []string {
	env := []string{}
	for _, key := range h.ForwardEnv {
		v, ok := os.LookupEnv(key)
		if ok {
			env = append(env, fmt.Sprintf("%s=%s", key, v))
		}
	}
	for key, value := range h.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// Interface guards
var (
	_ secrets.Handler = (*ExecHandler)(nil)
)
