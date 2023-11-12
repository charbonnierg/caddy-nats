// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

// Package client provides a struct which can be reused by libraries to
// declare options to connect to a NATS server.
// All public properties should be considered immutable, and should not be modified
// after creating the client.
// Options can be modified using the SetOptions method before the first connection.
package natsclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"go.uber.org/zap"
)

type InProcessConnProvider interface {
	GetServer() (*server.Server, error)
}

// NatsClient represents a connection to a JetStream enabled
// NATS server. An empty NatsClient is not usable, and must be
// provisioned AND connected before being used.
type NatsClient struct {
	sdk    *sdk
	opts   *nats.Options
	logger *zap.Logger
	ctx    caddy.Context
	done   chan struct{}

	provider InProcessConnProvider `json:"-"`

	Internal                       bool          `json:"internal,omitempty"`
	Name                           string        `json:"name,omitempty"`
	Servers                        []string      `json:"servers,omitempty"`
	Username                       string        `json:"username,omitempty"`
	Password                       string        `json:"password,omitempty"`
	Token                          string        `json:"token,omitempty"`
	Credentials                    string        `json:"credentials,omitempty"`
	Seed                           string        `json:"seed,omitempty"`
	Jwt                            string        `json:"jwt,omitempty"`
	JSDomain                       string        `json:"jetstream_domain,omitempty"`
	JSPrefix                       string        `json:"jetstream_prefix,omitempty"`
	InboxPrefix                    string        `json:"inbox_prefix,omitempty"`
	NoRandomize                    bool          `json:"no_randomize,omitempty"`
	DontReconnect                  bool          `json:"dont_reconnect,omitempty"`
	DontReconnectOnFirstConnection bool          `json:"dont_reconnect_on_first_connection,omitempty"`
	ReconnectWait                  time.Duration `json:"reconnect_wait,omitempty"`
	PingInterval                   time.Duration `json:"ping_interval,omitempty"`
}

// Provision must be called to initialize the client.
func (c *NatsClient) Provision(ctx caddy.Context, provider InProcessConnProvider) error {
	c.ctx = ctx
	c.logger = ctx.Logger().Named("client")
	c.provider = provider
	c.done = make(chan struct{})
	opts := nats.GetDefaultOptions()
	opts.AllowReconnect = !c.DontReconnect
	opts.RetryOnFailedConnect = !c.DontReconnectOnFirstConnection
	if c.Internal && c.Servers != nil {
		return errors.New("cannot specify both server and servers")
	}
	if c.Name != "" {
		opts.Name = c.Name
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
			return fmt.Errorf("failed to configure user credentials: %v", err)
		}
	}
	if c.Seed != "" {
		private, err := nkeys.FromSeed([]byte(c.Seed))
		if err != nil {
			return fmt.Errorf("failed to decode nkey seed: %v", err)
		}
		public, err := private.PublicKey()
		if err != nil {
			return err
		}
		if err := nats.Nkey(public, private.Sign)(&opts); err != nil {
			return fmt.Errorf("failed to configure public nkey and signature callback: %v", err)
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
	if c.ReconnectWait != 0 {
		opts.ReconnectWait = c.ReconnectWait
	}
	c.opts = &opts
	return nil
}

// SetOptions can be used to configure the NATS client before first connection.
func (c *NatsClient) SetOptions(opt ...nats.Option) error {
	if c.sdk != nil {
		return errors.New("cannot set options after first connection")
	}
	for _, o := range opt {
		err := o(c.opts)
		if err != nil {
			return err
		}
	}
	return nil
}

// Connect connects to the NATS server and returns a JetStream
// context. If the connection is already established, it returns
// the existing JetStream context.
func (c *NatsClient) Connect() error {
	if c.sdk != nil {
		return nil
	}
	// Override closed cb
	cb := c.opts.ClosedCB
	c.opts.ClosedCB = func(nc *nats.Conn) {
		defer func() {
			if c.done != nil {
				close(c.done)
			}
		}()
		if cb != nil {
			cb(nc)
		}
	}
	errorCb := c.opts.AsyncErrorCB
	c.opts.AsyncErrorCB = func(nc *nats.Conn, sub *nats.Subscription, err error) {
		if errorCb != nil {
			errorCb(nc, sub, err)
		}
		c.logger.Error("NATS connection error", zap.Error(err))
	}
	if err := c.validate(); err != nil {
		return err
	}
	// Override options if needed.
	if !c.Internal && c.Servers == nil {
		if c.provider == nil {
			c.opts.Servers = []string{"nats://localhost:4222"}
		} else {
			srv, err := c.provider.GetServer()
			if err != nil {
				return fmt.Errorf("failed to get inprocess server: %v", err)
			}
			c.opts.InProcessServer = srv
			c.opts.Servers = nil
		}
	} else if c.Internal {
		if c.provider == nil {
			return errors.New("internal client requires a server")
		}
		// If internal is True, a server must be provided in the context.
		srv, err := c.provider.GetServer()
		if err != nil {
			return fmt.Errorf("failed to get inprocess server: %v", err)
		}
		c.opts.InProcessServer = srv
		c.opts.Servers = nil
	} else {
		// Otherwise, use the provided servers.
		c.opts.Servers = c.Servers
	}
	// Validate options
	if err := c.validate(); err != nil {
		return err
	}
	// Connect
	nc, err := c.opts.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server: %v", err)
	}
	jsopts := []jetstream.JetStreamOpt{}
	jsctxopts := []nats.JSOpt{}
	var js jetstream.JetStream
	if c.JSPrefix != "" {
		jsctxopts = append(jsctxopts, nats.APIPrefix(c.JSPrefix))
		js, err = jetstream.NewWithAPIPrefix(nc, c.JSPrefix, jsopts...)
	} else if c.JSDomain != "" {
		jsctxopts = append(jsctxopts, nats.Domain(c.JSDomain))
		js, err = jetstream.NewWithDomain(nc, c.JSDomain, jsopts...)
	} else {
		js, err = jetstream.New(nc, jsopts...)
	}
	if err != nil {
		c.Close()
		return fmt.Errorf("invalid JetStream configuration: %v", err)
	}
	jsctx, err := nc.JetStream(jsctxopts...)
	if err != nil {
		c.Close()
		return fmt.Errorf("invalid JetStream context: %v", err)
	}
	connection := &sdk{
		nc:    nc,
		js:    js,
		jsctx: jsctx,
	}
	c.sdk = connection
	return nil
}

// Close closes the connection to the NATS server.
func (c *NatsClient) Close() error {
	switch {
	case c.sdk == nil:
		return nil
	case c.sdk.nc == nil:
		return nil
	case c.sdk.nc.IsReconnecting():
		c.sdk.nc.Close()
		return nil
	case c.sdk.nc.IsDraining():
		// If client is currently draining, wait for it to be done.
		<-c.done
		return nil
	case c.sdk.nc.IsConnected():
		// If client is connected, drain it.
		c.sdk.nc.Close()
		<-c.done
		return nil
	case !c.sdk.nc.IsClosed():
		// If client is not closed, close it.
		c.sdk.nc.Close()
		<-c.done
		return nil
	default:
		return nil
	}
}

func (c *NatsClient) JetStream() (jetstream.JetStream, error) {
	if c.sdk == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}
	return c.sdk.js, nil
}

func (c *NatsClient) JetStreamContext() (nats.JetStreamContext, error) {
	if c.sdk == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}
	return c.sdk.jsctx, nil
}

func (c *NatsClient) Nats() (*nats.Conn, error) {
	if c.sdk == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}
	return c.sdk.nc, nil
}

// validate checks that the client configuration is valid.
func (c *NatsClient) validate() error {
	if c.opts == nil {
		return errors.New("options have not been provisioned")
	}
	o := c.opts
	// Verify user/pass
	if o.User != "" && o.Token != "" {
		return errors.New("cannot specify both username and token")
	}
	if o.User != "" && o.Nkey != "" {
		return errors.New("cannot specify both username and nkey")
	}
	if o.User != "" && o.SignatureCB != nil {
		return errors.New("cannot specify both username and credentials")
	}
	if o.User != "" && o.UserJWT != nil {
		return errors.New("cannot specify both username and jwt")
	}
	if o.Password != "" && o.User == "" {
		return errors.New("cannot specify password without username")
	}
	// Verify token
	if o.Token != "" && o.SignatureCB != nil {
		return errors.New("cannot specify both token and credentials")
	}
	if o.Token != "" && o.Nkey != "" {
		return errors.New("cannot specify both token and nkey seed")
	}
	if c.Token != "" && c.Jwt != "" {
		return errors.New("cannot specify both token and jwt")
	}
	return nil
}

func (c *NatsClient) ConfigureStream(ctx context.Context, stream *Stream) error {
	if ctx == nil {
		ctx = c.ctx
	}
	return stream.Configure(ctx, c.sdk.JetStream())
}

func (c *NatsClient) ConfigureConsumer(ctx context.Context, consumer *Consumer) error {
	if ctx == nil {
		ctx = c.ctx
	}
	return consumer.Configure(ctx, c.sdk.JetStream())
}

func (c *NatsClient) ConfigureObjectStore(ctx context.Context, store *ObjectStore) error {
	if ctx == nil {
		ctx = c.ctx
	}
	return store.Configure(ctx, c.sdk.JetStreamContext())
}

func (c *NatsClient) ConfigureKeyValueStore(ctx context.Context, store *KeyValueStore) error {
	if ctx == nil {
		ctx = c.ctx
	}
	return store.Configure(ctx, c.sdk.JetStream())
}

func (c *NatsClient) ConfigureService(service ServiceProvider) (micro.Service, error) {
	definition, err := service.Definition()
	if err != nil {
		return nil, fmt.Errorf("failed to get service definition: %v", err)
	}
	srv, err := definition.Configure(c.sdk.Nats())
	if err != nil {
		return nil, fmt.Errorf("failed to start service: %v", err)
	}
	return srv, nil
}

// sdk is a struct holding the NATS connection and JetStream context.
// It is used to pass both the connection, and the JetStream context so that
// users do not have to worry about JetStream configuration such as domain or API prefix.
// Connections are already initialized and ready to use.
type sdk struct {
	nc    *nats.Conn
	js    jetstream.JetStream
	jsctx nats.JetStreamContext
}

// Nats returns the NATS connection (core API)
func (c *sdk) Nats() *nats.Conn {
	return c.nc
}

// JetStream returns the JetStream manager (recommended JetStream API)
func (c *sdk) JetStream() jetstream.JetStream {
	return c.js
}

// JetStreamContext returns the JetStream context (legacy JetStream API)
func (c *sdk) JetStreamContext() nats.JetStreamContext {
	return c.jsctx
}

func (c *NatsClient) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // skip "client"
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "in_process", "internal":
			if err := parser.ParseBool(d, &c.Internal); err != nil {
				return err
			}
		case "name":
			if err := parser.ParseString(d, &c.Name); err != nil {
				return err
			}
		case "servers":
			if err := parser.ParseStringArray(d, &c.Servers); err != nil {
				return err
			}
		case "username":
			if err := parser.ParseString(d, &c.Username); err != nil {
				return err
			}
		case "password":
			if err := parser.ParseString(d, &c.Password); err != nil {
				return err
			}
		case "token":
			if err := parser.ParseString(d, &c.Token); err != nil {
				return err
			}
		case "credentials":
			if err := parser.ParseString(d, &c.Credentials); err != nil {
				return err
			}
		case "seed":
			if err := parser.ParseString(d, &c.Seed); err != nil {
				return err
			}
		case "jwt":
			if err := parser.ParseString(d, &c.Jwt); err != nil {
				return err
			}
		case "jetstream_domain", "js_domain":
			if err := parser.ParseString(d, &c.JSDomain); err != nil {
				return err
			}
		case "jetstream_prefix", "jetstream_api_prefix", "js_prefix", "js_api_prefix":
			if err := parser.ParseString(d, &c.JSPrefix); err != nil {
				return err
			}
		case "inbox_prefix":
			if err := parser.ParseString(d, &c.InboxPrefix); err != nil {
				return err
			}
		case "no_randomize":
			if err := parser.ParseBool(d, &c.NoRandomize); err != nil {
				return err
			}
		case "ping_interval":
			if err := parser.ParseDuration(d, &c.PingInterval); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
