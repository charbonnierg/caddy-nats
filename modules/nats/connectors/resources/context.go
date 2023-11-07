package resources

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats-server/v2/server"
)

type clientsKey struct{}
type inProcessConnProviderKey struct{}

type inProcessConnProvider interface {
	GetServer() (*server.Server, error)
}

func SetInternalConnProviderInContext(ctx context.Context, provider inProcessConnProvider) context.Context {
	return context.WithValue(ctx, inProcessConnProviderKey{}, provider)
}

func SetInternalConnProviderInCaddyContext(ctx *caddy.Context, provider inProcessConnProvider) {
	ctx.Context = SetInternalConnProviderInContext(ctx.Context, provider)
}

func GetInProcessConnProviderFromContext(ctx context.Context) (inProcessConnProvider, bool) {
	raw := ctx.Value(inProcessConnProviderKey{})
	if raw == nil {
		return nil, false
	}
	provider, ok := raw.(inProcessConnProvider)
	if !ok {
		return nil, false
	}
	return provider, true
}

func SetClientsInContext(ctx context.Context, clients *Clients) context.Context {
	return context.WithValue(ctx, clientsKey{}, clients)
}

func SetClientsInCaddyContext(ctx *caddy.Context, clients *Clients) {
	ctx.Context = SetClientsInContext(ctx.Context, clients)
}

func GetClientsFromContext(ctx context.Context) (*Clients, bool) {
	raw := ctx.Value(clientsKey{})
	if raw == nil {
		return nil, false
	}
	clients, ok := raw.(*Clients)
	if !ok {
		return nil, false
	}
	return clients, true
}
