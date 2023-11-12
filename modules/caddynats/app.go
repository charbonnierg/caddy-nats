package caddynats

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/quara-dev/beyond"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(App{})
	httpcaddyfile.RegisterGlobalOption("nats_server", parseGlobalOption)
	httpcaddyfile.RegisterGlobalOption("broker", parseGlobalOption)
}

func Load(ctx caddy.Context) (*App, error) {
	unm, err := ctx.App("nats_server")
	if err != nil {
		return nil, err
	}
	appIface, ok := unm.(*App)
	if !ok {
		return nil, errors.New("not a valid nats app")
	}
	return appIface, nil
}

func ProvisionClientConnection(ctx caddy.Context, account string, client *natsclient.NatsClient) error {
	app, err := Load(ctx)
	if err != nil {
		return err
	}
	if err := app.ProvisionClientConnection(account, client); err != nil {
		return err
	}
	return nil
}

func ProvisionClientConnectionBeyond(ctx caddy.Context, account string, client *natsclient.NatsClient) error {
	unm, err := beyond.Load(ctx, "nats_server")
	if err != nil {
		return err
	}
	app, ok := unm.(*App)
	if !ok {
		return errors.New("not a valid nats app")
	}
	if err := app.ProvisionClientConnection(account, client); err != nil {
		return err
	}
	return nil
}

type App struct {
	*Server
}

func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server",
		New: func() caddy.Module { return new(App) },
	}
}

func (a *App) Provision(ctx caddy.Context) error {
	if _, err := beyond.Register(ctx, a); err != nil {
		return err
	}
	if err := a.Server.Provision(ctx); err != nil {
		return err
	}
	return nil
}

func (a *App) Validate() error {
	if a.Options != nil {
		a.logger.Debug("nats server adjusted config", zap.Any("options", a.Options))
	}
	return nil
}

func (a *App) Start() error {
	if err := a.Server.Start(); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop() error {
	if err := a.Server.Stop(); err != nil {
		return err
	}
	return nil
}

func (a *App) Context() caddy.Context {
	return a.ctx
}

func (a *App) Logger() *zap.Logger {
	return a.logger
}
