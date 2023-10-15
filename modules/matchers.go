// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
)

func init() {
	caddy.RegisterModule(ClientInfoMatcher{})
	caddy.RegisterModule(ConnectOptsMatcher{})
}

type Matcher interface {
	Match(request *jwt.AuthorizationRequestClaims) bool
}

type Matchers []Matcher

type ClientInfoMatcher struct {
	Host string `json:"host,omitempty"`
	User string `json:"user,omitempty"`
	Kind string `json:"kind,omitempty"`
	Type string `json:"type,omitempty"`
}

func (ClientInfoMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.client_info",
		New: func() caddy.Module { return new(ClientInfoMatcher) },
	}
}

func (m *ClientInfoMatcher) Match(request *jwt.AuthorizationRequestClaims) bool {
	if m.Host != "" && m.Host != request.ClientInformation.Host {
		return false
	}
	if m.User != "" && m.User != request.ClientInformation.User {
		return false
	}
	if m.Kind != "" && m.Kind != request.ClientInformation.Kind {
		return false
	}
	if m.Type != "" && m.Type != request.ClientInformation.Type {
		return false
	}
	return true
}

type ConnectOptsMatcher struct {
	Name     string `json:"name,omitempty"`
	User     string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Lang     string `json:"lang,omitempty"`
	Version  string `json:"version,omitempty"`
	Protocol int    `json:"protocol,omitempty"`
}

func (ConnectOptsMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.connect_opts",
		New: func() caddy.Module { return new(ConnectOptsMatcher) },
	}
}

func (c *ConnectOptsMatcher) Match(request *jwt.AuthorizationRequestClaims) bool {
	if c.User != "" && c.User != request.ConnectOptions.Username {
		return false
	}
	if c.Password != "" && c.Password != request.ConnectOptions.Password {
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
	_ Matcher = (*ClientInfoMatcher)(nil)
	_ Matcher = (*ConnectOptsMatcher)(nil)
)
