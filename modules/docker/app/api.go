package app

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

func (a *App) Logger() *zap.Logger { return a.logger }

func (a *App) Context() caddy.Context { return a.ctx }

func (a *App) Client() (*client.Client, error) {
	if a.client == nil {
		return nil, errors.New("docker client is not initialized")
	}
	return a.client.client, nil
}

func (a *App) Reconnect() (*client.Client, error) {
	if a.client == nil {
		return nil, errors.New("docker client is not initialized")
	}
	client, err := NewDockerClient(a.ctx, a.ClientOptions)
	if err != nil {
		return nil, err
	}
	a.client = client
	return client.client, nil
}
