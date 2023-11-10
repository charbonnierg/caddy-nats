// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package client_info

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats"
)

func init() {
	caddy.RegisterModule(ClientInfoMatcher{})
}

func (ClientInfoMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.matchers.client_info",
		New: func() caddy.Module { return new(ClientInfoMatcher) },
	}
}

// ClientInfoMatcher matches a connection according to its client information.
// The following fields can be used:
//
// - in_process: the connection must be in process when true
// - host: the connection must have the specified host
// - user: the connection must have the specified user
// - kind: the connection must have the specified kind (nats, websocket, ...)
// - type: the connection must have the specified type (client, leafnode, ...)
type ClientInfoMatcher struct {
	InProcess bool   `json:"in_process,omitempty"`
	Host      string `json:"host,omitempty"`
	User      string `json:"user,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Type      string `json:"type,omitempty"`
}

func (m *ClientInfoMatcher) Provision(app nats.App) error {
	return nil
}

func (m *ClientInfoMatcher) Match(request *jwt.AuthorizationRequestClaims) bool {
	if m.InProcess && request.ClientInformation.Host != "" {
		return false
	}
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

var (
	_ nats.Matcher = (*ClientInfoMatcher)(nil)
)
