package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/quara-dev/beyond/modules/docker"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(DockerUpstreams{})
}

type DockerUpstreams struct {
	app      docker.App
	client   *client.Client
	ctx      caddy.Context
	mutex    *sync.Mutex
	logger   *zap.Logger
	upstream *reverseproxy.Upstream
	cancel   context.CancelFunc

	Container string `json:"container,omitempty"`
	Network   string `json:"network,omitempty"`
	Port      int    `json:"port,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (DockerUpstreams) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.reverse_proxy.upstreams.docker",
		New: func() caddy.Module { return new(DockerUpstreams) },
	}
}

func (mu *DockerUpstreams) Provision(ctx caddy.Context) error {
	mu.ctx = ctx
	mu.mutex = new(sync.Mutex)
	mu.logger = ctx.Logger()
	app, err := docker.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading docker app module: %v", err)
	}
	mu.app = app
	return nil
}

func (mu *DockerUpstreams) watch() error {
	mu.mutex.Lock()
	defer mu.mutex.Unlock()
	if mu.cancel != nil {
		return nil
	}
	ctx, cancel := context.WithCancel(mu.ctx)
	mu.cancel = cancel
	client, err := mu.app.Client()
	if err != nil {
		return fmt.Errorf("getting docker client: %v", err)
	}
	mu.client = client
	container, err := mu.client.ContainerInspect(ctx, mu.Container)
	if err != nil {
		if strings.Contains(err.Error(), "context cancelled") {
			return nil
		}
		return fmt.Errorf("inspecting container: %v", err)
	}
	var addr string = container.NetworkSettings.IPAddress
	if mu.Network != "" {
		for name, network := range container.NetworkSettings.Networks {
			if name == mu.Network || network.NetworkID == mu.Network {
				if network.IPAMConfig != nil && network.IPAMConfig.IPv4Address != "" {
					addr = network.IPAMConfig.IPv4Address
				} else if network.IPAddress != "" {
					addr = network.IPAddress
				}
				break
			}
			for _, alias := range network.Aliases {
				if alias == mu.Network {
					if network.IPAMConfig != nil && network.IPAMConfig.IPv4Address != "" {
						addr = network.IPAMConfig.IPv4Address
					} else if network.IPAddress != "" {
						addr = network.IPAddress
					}
					break
				}
			}
			if addr != "" {
				break
			}
		}
	}
	if addr == "" {
		mu.logger.Error("did not found ip address", zap.Any("container", container))
		return fmt.Errorf("container not found")
	}
	// Set upstream
	mu.upstream = &reverseproxy.Upstream{
		Dial: fmt.Sprintf("%s:%d", addr, mu.Port),
	}
	// Subscribe to events
	msgs, errs := client.Events(ctx, types.EventsOptions{
		Since: time.Now().Format(time.RFC3339),
		Filters: filters.NewArgs(
			filters.Arg("container", mu.Container),
		),
	})
	// Process events within goroutine
	go func() {
		defer func() {
			mu.cancel()
			mu.cancel = nil
		}()
		mu.logger.Info("docker event listener started")
		for {
			select {
			case <-ctx.Done():
				mu.logger.Info("docker event listener stopped")
				return
			case err := <-errs:
				if err.Error() == "unexpected EOF" {
					mu.logger.Warn("docker client disconnected, attempting to reconnect")
					client, err = mu.app.Reconnect()
					if err != nil {
						mu.logger.Warn("failed to reconnect docker client", zap.Error(err))
					}
					mu.logger.Info("docker client reconnected successfully")
					return
				}
				mu.logger.Error("docker event error", zap.Error(err))
				mu.upstream = nil
				return
			case msg := <-msgs:
				switch msg.Action {
				case "die", "stop", "kill":
					mu.upstream = nil
				default:
					mu.logger.Warn("docker event", zap.String("action", msg.Action))
					container, err := mu.client.ContainerInspect(ctx, mu.Container)
					if err != nil {
						if strings.Contains(err.Error(), "context cancelled") {
							return
						}
						mu.logger.Error("inspecting container", zap.Error(err))
						mu.upstream = nil
						continue
					}
					var addr string = container.NetworkSettings.IPAddress
					if mu.Network != "" {
						for _, network := range container.NetworkSettings.Networks {
							if network.NetworkID == mu.Network {
								if network.IPAMConfig != nil && network.IPAMConfig.IPv4Address != "" {
									addr = network.IPAMConfig.IPv4Address
								} else if network.IPAddress != "" {
									addr = network.IPAddress
								}
								break
							}
							for _, alias := range network.Aliases {
								if alias == mu.Network {
									if network.IPAMConfig != nil && network.IPAMConfig.IPv4Address != "" {
										addr = network.IPAMConfig.IPv4Address
									} else if network.IPAddress != "" {
										addr = network.IPAddress
									}
									break
								}
							}
						}
					}
					mu.upstream = &reverseproxy.Upstream{
						Dial: fmt.Sprintf("%s:%d", addr, mu.Port),
					}
				}
			}
		}
	}()
	return nil
}

func (mu DockerUpstreams) GetUpstreams(r *http.Request) ([]*reverseproxy.Upstream, error) {
	if err := mu.watch(); err != nil {
		return nil, fmt.Errorf("watching container: %v", err)
	}
	if mu.upstream == nil {
		return nil, fmt.Errorf("upstream not found")
	}
	return []*reverseproxy.Upstream{mu.upstream}, nil
}

func (mu *DockerUpstreams) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if err := parser.ExpectString(d); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "container":
			if err := parser.ParseString(d, &mu.Container); err != nil {
				return err
			}
			mu.Container = d.Val()
		case "network":
			if err := parser.ParseString(d, &mu.Network); err != nil {
				return err
			}
		case "port":
			if err := parser.ParseNetworkPort(d, &mu.Port); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
