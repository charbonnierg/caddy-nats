// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

// NatsClient represents a connection to a JetStream enabled
// NATS server. It is lazy and will only connect when
// the first time Connect method is called.
type NatsClient struct {
	closed      bool
	connections *Clients
	ctx         context.Context

	Internal     bool          `json:"internal,omitempty"`
	Name         string        `json:"name,omitempty"`
	Servers      []string      `json:"servers,omitempty"`
	Username     string        `json:"username,omitempty"`
	Password     string        `json:"password,omitempty"`
	Token        string        `json:"token,omitempty"`
	Credentials  string        `json:"credentials,omitempty"`
	Seed         string        `json:"seed,omitempty"`
	Jwt          string        `json:"jwt,omitempty"`
	JSDomain     string        `json:"jetstream_domain,omitempty"`
	JSPrefix     string        `json:"jetstream_prefix,omitempty"`
	InboxPrefix  string        `json:"inbox_prefix,omitempty"`
	NoRandomize  bool          `json:"no_randomize,omitempty"`
	PingInterval time.Duration `json:"ping_interval,omitempty"`
}

func (c *NatsClient) Provision(ctx context.Context) error {
	if c.closed {
		return errors.New("client is closed")
	}
	c.ctx = ctx
	if err := c.validate(); err != nil {
		return err
	}
	return nil
}

// Connect connects to the NATS server and returns a JetStream
// context. If the connection is already established, it returns
// the existing JetStream context.
func (c *NatsClient) Connect() (*Clients, error) {
	if c.closed {
		return nil, errors.New("client is closed")
	}
	if c.connections != nil {
		return c.connections, nil
	}
	opts := nats.GetDefaultOptions()
	opts.AllowReconnect = true
	opts.RetryOnFailedConnect = true
	opts.ReconnectWait = 100 * time.Millisecond
	if c.Name != "" {
		opts.Name = c.Name
	}
	if !c.Internal && c.Servers == nil {
		// If both internal is false and servers is nil, lookup in context to see
		// if there is an internal server. If so, use it. Otherwise, use default
		// server URL (nats://localhost:4222).
		srv, ok := GetInProcessConnProviderFromContext(c.ctx)
		if !ok {
			opts.Servers = []string{"nats://localhost:4222"}
		} else {
			server, err := srv.GetServer()
			if err != nil {
				return nil, fmt.Errorf("failed to get inprocess server: %v", err)
			}
			opts.InProcessServer = server
			opts.Servers = nil
		}
	} else if c.Internal {
		// If internal is True, a server must be provided in the context.
		srv, ok := GetInProcessConnProviderFromContext(c.ctx)
		if !ok {
			return nil, errors.New("failed to get inprocess server from context")
		}
		server, err := srv.GetServer()
		if err != nil {
			return nil, fmt.Errorf("failed to get inprocess server: %v", err)
		}
		opts.InProcessServer = server
		opts.Servers = nil
	} else {
		// Otherwise, use the provided servers.
		opts.Servers = c.Servers
	}
	if c.Username != "" {
		opts.User = c.Username
	}
	if c.Password != "" {
		opts.Password = c.Password
	}
	if c.Token != "" {
		opts.Token = c.Token
	}
	if c.Credentials != "" {
		if err := nats.UserCredentials(c.Credentials)(&opts); err != nil {
			return nil, fmt.Errorf("failed to configure user credentials: %v", err)
		}
	}
	if c.Seed != "" {
		private, err := nkeys.FromSeed([]byte(c.Seed))
		if err != nil {
			return nil, fmt.Errorf("failed to decode nkey seed: %v", err)
		}
		public, err := private.PublicKey()
		if err != nil {
			return nil, err
		}
		if err := nats.Nkey(public, private.Sign)(&opts); err != nil {
			return nil, fmt.Errorf("failed to configure public nkey and signature callback: %v", err)
		}
	}
	if c.Jwt != "" {
		opts.UserJWT = func() (string, error) { return c.Jwt, nil }
	}
	if c.InboxPrefix != "" {
		opts.InboxPrefix = c.InboxPrefix
	}
	if c.NoRandomize {
		opts.NoRandomize = true
	}
	if c.PingInterval != 0 {
		opts.PingInterval = c.PingInterval
	}
	nc, err := opts.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %v", err)
	}
	jsopts := []nats.JSOpt{}
	if c.JSPrefix != "" {
		jsopts = append(jsopts, nats.APIPrefix(c.JSPrefix))
	}
	if c.JSDomain != "" {
		jsopts = append(jsopts, nats.Domain(c.JSDomain))
	}
	js, err := nc.JetStream(jsopts...)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("invalid JetStream configuration: %v", err)
	}
	connection := &Clients{
		nc: nc,
		js: js,
	}
	c.connections = connection
	return connection, nil
}

// Close closes the connection to the NATS server.
func (c *NatsClient) Close() {
	defer func() {
		c.closed = true
		c.connections = nil
	}()
	if c.connections == nil {
		return
	}
	if c.connections.nc == nil {
		return
	}
	nc := c.connections.nc
	if !nc.IsClosed() && !nc.IsDraining() {
		nc.Close()
	}
}

// validate checks that the client configuration is valid.
func (c *NatsClient) validate() error {
	if c.Internal && c.Servers != nil {
		return errors.New("cannot specify both server and servers")
	}
	// Verify user/pass
	if c.Username != "" && c.Token != "" {
		return errors.New("cannot specify both username and token")
	}
	if c.Username != "" && c.Credentials != "" {
		return errors.New("cannot specify both username and credentials")
	}
	if c.Username != "" && c.Seed != "" {
		return errors.New("cannot specify both username and nkey seed")
	}
	if c.Username != "" && c.Jwt != "" {
		return errors.New("cannot specify both username and jwt")
	}
	if c.Password != "" && c.Username == "" {
		return errors.New("cannot specify password without username")
	}
	// Verify token
	if c.Token != "" && c.Credentials != "" {
		return errors.New("cannot specify both token and credentials")
	}
	if c.Token != "" && c.Seed != "" {
		return errors.New("cannot specify both token and nkey seed")
	}
	if c.Token != "" && c.Jwt != "" {
		return errors.New("cannot specify both token and jwt")
	}
	// Verify credentials
	if c.Credentials != "" && c.Seed != "" {
		return errors.New("cannot specify both credentials and nkey seed")
	}
	if c.Credentials != "" && c.Jwt != "" {
		return errors.New("cannot specify both credentials and jwt")
	}
	return nil
}

type Clients struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func (c *Clients) Core() *nats.Conn {
	return c.nc
}

func (c *Clients) JetStream() nats.JetStreamContext {
	return c.js
}
