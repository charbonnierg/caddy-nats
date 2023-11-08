package app

import (
	"fmt"
	"os"
	"os/exec"

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
	p.logger = app.Logger().Named(p.Name)
	curPath, ok := p.Environment["PATH"]
	if !ok {
		curPath = os.Getenv("PATH")
	}
	if p.VirtualEnv != "" {
		p.Environment = fnutils.DefaultIfEmptyMap(p.Environment, map[string]string{})
		p.Environment["VIRTUAL_ENV"] = p.VirtualEnv
		p.Environment["PATH"] = fmt.Sprintf("%s/bin:%s", p.VirtualEnv, curPath)
		if p.Command == "python" || p.Command == "python3" {
			p.Command = fmt.Sprintf("%s/bin/%s", p.VirtualEnv, p.Command)
		}
	}
	cmd := exec.Command(p.Command, p.Args...)
	cmd.Env = p.gatherEnvironmentVariables()
	cmd.Dir = p.WorkingDir
	if p.ForwardStderr {
		cmd.Stderr = os.Stderr
	}
	if p.ForwardStdout {
		cmd.Stdout = os.Stdout
	}
	p.cmd = cmd
	return nil
}

func (p *PythonProcess) Start() error {
	p.logger.Info("starting python process", zap.String("name", p.Name), zap.Any("command", p.cmd))
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	p.logger.Info("started python process", zap.String("name", p.Name), zap.Any("pid", p.cmd.Process.Pid))
	go func() {
		p.logger.Info("starting process monitor", zap.String("name", p.Name), zap.Int("pid", p.cmd.Process.Pid))
		if err := p.cmd.Wait(); err != nil {
			p.logger.Error("python process exited", zap.String("name", p.Name), zap.Error(err))
		}
		p.logger.Info("python process exited without error", zap.String("name", p.Name))
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
	p.logger.Info("stopping python process", zap.String("name", p.Name))
	if err := p.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

func (p *PythonProcess) gatherEnvironmentVariables() []string {
	env := []string{}
	for key, value := range p.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}
