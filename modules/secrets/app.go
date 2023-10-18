package secrets

import (
	"fmt"

	"github.com/caddyserver/caddy/v2"
	"github.com/charbonnierg/beyond"
	docker "github.com/charbonnierg/beyond/modules/docker/interfaces"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(new(App))
}

type App struct {
	ctx    caddy.Context
	logger *zap.Logger
	beyond *beyond.Beyond
	docker docker.DockerApp
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "secrets",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	a.ctx = ctx
	a.logger = ctx.Logger()

	// Anything that MUST be done before allowing other apps to access the secret app
	// should happend BEFORE loading the beyond module.
	// E.G., here:
	// ...

	// This will load the beyond module and register the "secrets" app within beyond module
	if err := a.register(); err != nil {
		return err
	}
	// At this point we can use the beyond module to load other apps
	// Let's load the secret app
	unm, err := a.beyond.LoadApp(a, "docker")
	if err != nil {
		return fmt.Errorf("failed to load docker app: %v", err)
	}
	a.docker = unm.(docker.DockerApp)
	return nil
}

// Helper function to load the beyond module and register the "secrets" app within beyond module
func (a *App) register() error {
	b, err := beyond.RegisterApp(a.ctx, a)
	if err != nil {
		return err
	}
	a.beyond = b
	return nil
}

func (a *App) Start() error {
	a.logger.Info("Starting secrets app")
	return nil
}

func (a *App) Stop() error {
	a.logger.Info("Stopping secrets app")
	return nil
}

var (
	_ caddy.App         = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
)
