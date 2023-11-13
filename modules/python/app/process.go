// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/python"
	"github.com/quara-dev/beyond/pkg/fnutils"
	"go.uber.org/zap"
)

type PythonProcess struct {
	Name          string            `json:"name,omitempty"`
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	VirtualEnv    string            `json:"virtualenv,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	ForwardStdout bool              `json:"forward_stdout,omitempty"`
	ForwardStderr bool              `json:"forward_stderr,omitempty"`

	logger *zap.Logger
	cmd    *exec.Cmd
}

func (p *PythonProcess) Provision(app python.App) error {
	repl := app.Replacer()
	command, err := repl.ReplaceOrErr(p.Command, true, true)
	if err != nil {
		return fmt.Errorf("failed to replace command: %w", err)
	}
	args := []string{}
	for _, arg := range p.Args {
		arg, err := repl.ReplaceOrErr(arg, true, true)
		if err != nil {
			return fmt.Errorf("failed to replace arg: %w", err)
		}
		args = append(args, arg)
	}
	workdir, err := repl.ReplaceOrErr(p.WorkingDir, true, true)
	if err != nil {
		return fmt.Errorf("failed to replace working_dir: %w", err)
	}
	curPath, ok := p.Environment["PATH"]
	if !ok {
		curPath = os.Getenv("PATH")
	}
	if p.VirtualEnv != "" {
		p.Environment = fnutils.DefaultIfEmptyMap(p.Environment, map[string]string{})
		p.Environment["VIRTUAL_ENV"] = p.VirtualEnv
		p.Environment["PATH"] = fmt.Sprintf("%s/bin:%s", p.VirtualEnv, curPath)
		if command == "python" || command == "python3" {
			command = fmt.Sprintf("%s/bin/%s", p.VirtualEnv, command)
		}
	}
	env, err := p.gatherEnvironmentVariables(repl)
	if err != nil {
		return fmt.Errorf("failed to gather environment variables: %w", err)
	}
	cmd := exec.Command(command, args...)
	cmd.Env = env
	cmd.Dir = workdir
	if p.ForwardStderr {
		cmd.Stderr = os.Stderr
	}
	if p.ForwardStdout {
		cmd.Stdout = os.Stdout
	}
	p.cmd = cmd
	p.logger = app.Logger().Named("process").With(zap.String("name", p.Name), zap.String("command", p.Command))
	return nil
}

func (p *PythonProcess) Start() error {
	p.logger.Info("starting python process")
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	p.logger.Info("started python process")
	go func() {
		p.logger.Info("starting process monitor", zap.Int("pid", p.cmd.Process.Pid))
		if err := p.cmd.Wait(); err != nil {
			p.logger.Error("python process exited", zap.Int("pid", p.cmd.Process.Pid), zap.Error(err))
		}
		p.logger.Info("python process exited without error", zap.Int("pid", p.cmd.Process.Pid))
	}()
	return nil
}

func (p *PythonProcess) Stop() error {
	if p.cmd == nil {
		return nil
	}
	if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		return nil
	}
	p.logger.Info("stopping python process")
	if err := p.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

func (p *PythonProcess) gatherEnvironmentVariables(repl *caddy.Replacer) ([]string, error) {
	env := []string{}
	for key, value := range p.Environment {
		value, err := repl.ReplaceOrErr(value, true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to replace environment variable: %w", err)
		}
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env, nil
}
