// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
	"github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/auth/template"
)

// AuthorizationRequest is the authorization request.
type AuthorizationRequest struct {
	claims  *jwt.AuthorizationRequestClaims
	context context.Context
}

// Claims returns the claims of the authorization request.
func (r *AuthorizationRequest) Claims() *jwt.AuthorizationRequestClaims {
	return r.claims
}

// Context returns the context of the authorization request.
func (r *AuthorizationRequest) Context() context.Context {
	return r.context
}

// Replacer returns the caddy replacer of the authorization request.
func (r *AuthorizationRequest) Replacer() *caddy.Replacer {
	return r.getReplacer()
}

func (r *AuthorizationRequest) setReplacer() *caddy.Replacer {
	repl := caddy.NewReplacer()
	template.AddSecretsVarsToReplacer(repl)
	template.AddAuthRequestVarsToReplacer(repl, r.claims)
	r.context = context.WithValue(r.context, template.ReplacerCtxKey{}, repl)
	return repl
}

func (r *AuthorizationRequest) getReplacer() *caddy.Replacer {
	raw := r.context.Value(template.ReplacerCtxKey{})
	if raw == nil {
		return r.setReplacer()
	}
	repl, ok := raw.(*caddy.Replacer)
	if !ok {
		return r.setReplacer()
	}
	return repl
}

var (
	_ nats.AuthRequest = (*AuthorizationRequest)(nil)
)
