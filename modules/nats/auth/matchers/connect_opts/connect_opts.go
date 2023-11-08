// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package connect_opts

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats"
)

func init() {
	caddy.RegisterModule(ConnectOptsMatcher{})
}

func (ConnectOptsMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.connect_opts",
		New: func() caddy.Module { return new(ConnectOptsMatcher) },
	}
}

// ConnectOptsMatcher matches a connection according to its connect options.
// The following fields can be used:
//
// - name: the connection must have the specified name
// - token: the connection must have the specified token
// - user: the connection must have the specified user
// - password: the connection must have the specified password
// - lang: the connection must have the specified lang
// - version: the connection must have the specified version
// - protocol: the connection must have the specified protocol
type ConnectOptsMatcher struct {
	Name     string `json:"name,omitempty"`
	Token    string `json:"token,omitempty"`
	User     string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Lang     string `json:"lang,omitempty"`
	Version  string `json:"version,omitempty"`
	Protocol int    `json:"protocol,omitempty"`
}

func (c *ConnectOptsMatcher) Match(request *jwt.AuthorizationRequestClaims) bool {
	if c.User != "" && c.User != request.ConnectOptions.Username {
		return false
	}
	if c.Password != "" && c.Password != request.ConnectOptions.Password {
		return false
	}
	if c.Token != "" && c.Token != request.ConnectOptions.Token {
		return false
	}
	if c.Lang != "" && c.Lang != request.ConnectOptions.Lang {
		return false
	}
	if c.Version != "" && c.Version != request.ConnectOptions.Version {
		return false
	}
	if c.Protocol != 0 && c.Protocol != request.ConnectOptions.Protocol {
		return false
	}
	return true
}

var (
	_ nats.Matcher = (*ConnectOptsMatcher)(nil)
)
