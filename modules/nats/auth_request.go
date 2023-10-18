package nats

import (
	"context"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/jwt/v2"
)

type AuthorizationRequest struct {
	Claims  *jwt.AuthorizationRequestClaims
	Context context.Context
}

func (r *AuthorizationRequest) setReplacer() *caddy.Replacer {
	repl := caddy.NewReplacer()
	AddSecretsVarsToReplacer(repl)
	AddAuthRequestVarsToReplacer(repl, r.Claims)
	r.Context = context.WithValue(r.Context, ReplacerCtxKey{}, repl)
	return repl
}

func (r *AuthorizationRequest) GetReplacer() *caddy.Replacer {
	raw := r.Context.Value(ReplacerCtxKey{})
	if raw == nil {
		return r.setReplacer()
	}
	repl, ok := raw.(*caddy.Replacer)
	if !ok {
		return r.setReplacer()
	}
	return repl
}

// ReplaceKnown is like ReplaceAll but only replaces placeholders that are known (recognized).
// Unrecognized placeholders will remain in the output.
func (r *AuthorizationRequest) ReplaceKnown(input string, empty string) string {
	return r.GetReplacer().ReplaceKnown(input, empty)
}

// ReplaceAll replaces all placeholders in input with their values.
func (r *AuthorizationRequest) ReplaceAll(input string, empty string) string {
	return r.GetReplacer().ReplaceAll(input, empty)
}

// AddReplacerMapper adds a replacer key and mapper function to the replacer.
// The function will be used to map keys to values when the replacer is used.
func (r *AuthorizationRequest) AddReplacerMapper(mapper func(key string) (any, bool)) {
	repl := r.GetReplacer()
	repl.Map(mapper)
}
