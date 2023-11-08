// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package client provides a struct which can be reused by libraries to
// declare options to connect to a NATS server.
// When developing a caddy module, you can use the Connection struct
// which embeds the NatsClient struct.
package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

type InProcessConnProvider interface {
	GetServer() (*server.Server, error)
}

// NatsClient represents a connection to a JetStream enabled
// NATS server. It is lazy and will only connect when
// the first time Connect method is called.
type NatsClient struct {
	closed      bool
	connections *Clients

	Provider InProcessConnProvider `json:"-"`

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
		if c.Provider == nil {
			opts.Servers = []string{"nats://localhost:4222"}
		} else {
			srv, err := c.Provider.GetServer()
			if err != nil {
				return nil, fmt.Errorf("failed to get inprocess server: %v", err)
			}
			opts.InProcessServer = srv
			opts.Servers = nil
		}
	} else if c.Internal {
		if c.Provider == nil {
			return nil, errors.New("internal client requires a server")
		}
		// If internal is True, a server must be provided in the context.
		srv, err := c.Provider.GetServer()
		if err != nil {
			return nil, fmt.Errorf("failed to get inprocess server: %v", err)
		}
		opts.InProcessServer = srv
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
func (c *NatsClient) Validate() error {
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

// Clients is a struct holding the NATS connection and JetStream context.
// It is used to pass both the connection, and the JetStream context so that
// users do not have to worry about JetStream configuration such as domain or API prefix.
// Connections are already initialized and ready to use.
type Clients struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// Core returns the NATS connection.
func (c *Clients) Core() *nats.Conn {
	return c.nc
}

// JetStream returns the JetStream context.
func (c *Clients) JetStream() nats.JetStreamContext {
	return c.js
}
