// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package oauth2app

import (
	"net/http"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"go.uber.org/zap"
)

// nextKey is a struct used as key to store the next handler in the request context.
// The go context module docs recommends to use a private struct rather than a string to avoid
// collisions with other packages.
type nextKey struct{}

// upstream is a struct that implements the http.Handler interface.
// It is called by oauth2-proxy gorilla mux when the request is authorized.
// It fetches the next handler from the request context and calls it.
// This whole thing relies on Endpoint.ServeHTTP to set the next handler in the request context
// under the nextKey{} key.
type upstream struct {
	sessionLoader func(r *http.Request) (*sessions.SessionState, error)
	logger        *zap.Logger
}

// setSessionLoader sets the session loader function.
// this is needed to avoid circular dependencies because proxy needs the upstream to be created
// but upstream need sessionLoader from proxy. Instead of passing the sessionLoader when creating
// the upstream, we set it after both the upstream and the proxy are created.
func (h *upstream) setSessionLoader(loader func(r *http.Request) (*sessions.SessionState, error)) {
	h.sessionLoader = loader
}

// ServeHTTP fetches the next handler from the request context and calls it.
// It is called as the upstream handler only when the request is authorized.
// It is called by oauth2-proxy gorilla mux, not by caddy.
// Since oauth2-proxy gorilla mux has no concept of "next handler", we must get the next handler
// from the request context. (as a reminder, next handler is set in context by Endpoint.ServeHTTP method)
func (h upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := h.sessionLoader(r)
	if err != nil {
		h.logger.Error("not authorized", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("serving authorized request", zap.String("email", session.Email), zap.String("expires_on", session.ExpiresOn.String()))
	nextRaw := r.Context().Value(nextKey{})
	if nextRaw == nil {
		h.logger.Error("next handler not found in request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	next, ok := nextRaw.(caddyhttp.Handler)
	if !ok {
		h.logger.Error("next handler is not an http.Handler")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := next.ServeHTTP(w, r); err != nil {
		h.logger.Error("error serving next handler", zap.Error(err))
	}
}
