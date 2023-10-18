package jetstream

import (
	"errors"
	"fmt"
	"time"

	natsapp "github.com/charbonnierg/beyond/modules/nats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

// Client represents a connection to a JetStream enabled
// NATS server. It is lazy and will only connect when
// the first time Connect method is called.
type Client struct {
	closed       bool
	server       natsapp.ServerProvider
	nc           *nats.Conn
	js           nats.JetStreamContext
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

func (c *Client) SetInProcessServerProvider(provider natsapp.ServerProvider) {
	c.server = provider
}

// Connect connects to the NATS server and returns a JetStream
// context. If the connection is already established, it returns
// the existing JetStream context.
func (c *Client) Connect() (nats.JetStreamContext, error) {
	if c.closed {
		return nil, errors.New("client is closed")
	}
	if c.js != nil {
		return c.js, nil
	}
	c.validate()
	opts := nats.GetDefaultOptions()
	opts.AllowReconnect = true
	opts.ReconnectWait = 100 * time.Millisecond
	if c.Name != "" {
		opts.Name = c.Name
	}
	if c.server != nil {
		srv, err := c.server.GetServer()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-process server: %v", err)
		}
		if err := nats.InProcessServer(srv)(&opts); err != nil {
			return nil, fmt.Errorf("failed to configure in-process server: %v", err)
		}
	}
	if c.Servers != nil {
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
	c.nc = nc
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
	c.js = js
	return js, nil
}

// Close closes the connection to the NATS server.
func (c *Client) Close() {
	if c.nc != nil && !c.nc.IsClosed() && !c.nc.IsDraining() {
		c.nc.Close()
	}
	c.nc = nil
	c.js = nil
	c.closed = true
}

// validate checks that the client configuration is valid.
func (c *Client) validate() error {
	if c.server != nil && c.Servers != nil {
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
