// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream_kv_get

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamKeyValueGet{})
	httpcaddyfile.RegisterHandlerDirective("jetstream_kv_get", parseHandlerDirective)
}

// CaddyModule implements the caddy.Module interface.
// It returns information about the caddy module.
func (JetStreamKeyValueGet) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.jetstream_kv_get",
		New: func() caddy.Module { return new(JetStreamKeyValueGet) },
	}
}

// JetStreamKeyValueGet is an http middleware which publishes incoming HTTP
// requests as NATS messages and wait for an acknowledgement to be
// sent by JetStream engine (notifying that message has been persisted into a stream)
// before replying with a 204 empty response.
// HTTP headers are included in the response indicating
// which stream the message was persisted into, as well as
// the sequence the message was inserted to.
type JetStreamKeyValueGet struct {
	ctx     caddy.Context
	logger  *zap.Logger
	bucket  jetstream.KeyValue
	Client  *natsclient.NatsClient `json:"client,omitempty"`
	Account string                 `json:"account,omitempty"`
	Bucket  string                 `json:"bucket,omitempty"`
	Key     string                 `json:"key,omitempty"`
}

// Provision implements the caddy.Provisioner interface.
// It is executed when module is loaded (on caddy startup,
// or on config reload), and is responsible for loading
// the nats.App module app in order to create a new
// nats.ClientConnection for this handler.
func (p *JetStreamKeyValueGet) Provision(ctx caddy.Context) error {
	p.ctx = ctx
	p.logger = ctx.Logger()
	if p.Client == nil {
		p.Client = &natsclient.NatsClient{Internal: true}
	}
	if err := caddynats.ProvisionClientConnection(ctx, p.Account, p.Client); err != nil {
		return err
	}
	return nil
}

// Helper function to send a 500 Internal Server Error
func (p JetStreamKeyValueGet) writeServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Internal server error"))
	w.WriteHeader(http.StatusInternalServerError)
}

func (p *JetStreamKeyValueGet) getBucket(name string) (jetstream.KeyValue, error) {
	if p.bucket != nil {
		return p.bucket, nil
	}
	js, err := p.Client.JetStream()
	if err != nil {
		return p.bucket, err
	}
	bucket, err := js.KeyValue(p.ctx, name)
	if err != nil {
		return p.bucket, err
	}
	p.bucket = bucket
	return p.bucket, nil
}

// ServeHTTP receives incoming HTTP requests and is responsible
// for publishing request as a NATS message before writing to
// the HTTP response writer.
func (p JetStreamKeyValueGet) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	bucket, err := p.getBucket(p.Bucket)
	if err != nil {
		p.logger.Error("Error getting NATS connection", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	entry, err := bucket.Get(p.ctx, p.Key)
	switch err {
	case jetstream.ErrKeyNotFound:
		w.WriteHeader(http.StatusNotFound)
		return nil
	case nil:
		break
	default:
		p.logger.Error("Error fetching message from stream", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	w.Header().Add("Nats-Js-Revision", fmt.Sprintf("%d", entry.Revision()))
	w.Header().Add("Nats-Js-Delta", fmt.Sprintf("%d", entry.Delta()))
	w.Header().Add("Nats-Js-Operation", fmt.Sprintf("%d", entry.Operation()))
	w.Write(entry.Value())
	w.WriteHeader(http.StatusOK)
	p.logger.Warn("HTTP request processed", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	return next.ServeHTTP(w, r)
}
