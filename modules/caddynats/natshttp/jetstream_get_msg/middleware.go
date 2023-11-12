// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream_get_msg

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
	caddy.RegisterModule(JetStreamGetMsg{})
	httpcaddyfile.RegisterHandlerDirective("jetstream_get_msg", parseHandlerDirective)
}

// CaddyModule implements the caddy.Module interface.
// It returns information about the caddy module.
func (JetStreamGetMsg) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.jetstream_get_msg",
		New: func() caddy.Module { return new(JetStreamGetMsg) },
	}
}

// JetStreamGetMsg is an http middleware which publishes incoming HTTP
// requests as NATS messages and wait for an acknowledgement to be
// sent by JetStream engine (notifying that message has been persisted into a stream)
// before replying with a 204 empty response.
// HTTP headers are included in the response indicating
// which stream the message was persisted into, as well as
// the sequence the message was inserted to.
type JetStreamGetMsg struct {
	ctx      caddy.Context
	logger   *zap.Logger
	stream   jetstream.Stream
	Client   *natsclient.NatsClient `json:"client,omitempty"`
	Account  string                 `json:"account,omitempty"`
	Stream   string                 `json:"stream,omitempty"`
	Subject  string                 `json:"subject,omitempty"`
	Sequence int                    `json:"sequence,omitempty"`
}

// Provision implements the caddy.Provisioner interface.
// It is executed when module is loaded (on caddy startup,
// or on config reload), and is responsible for loading
// the nats.App module app in order to create a new
// nats.ClientConnection for this handler.
func (p *JetStreamGetMsg) Provision(ctx caddy.Context) error {
	p.ctx = ctx
	p.logger = ctx.Logger()
	if p.Client == nil {
		p.Client = &natsclient.NatsClient{Internal: true}
	}
	if p.Stream == "" && p.Subject == "" {
		return fmt.Errorf("stream or subject must be specified")
	}
	if p.Subject != "" && p.Sequence != 0 {
		return fmt.Errorf("sequence is mutually exclusive with subject")
	}
	if err := caddynats.ProvisionClientConnection(ctx, p.Account, p.Client); err != nil {
		return err
	}
	return nil
}

// Helper function to send a 500 Internal Server Error
func (p JetStreamGetMsg) writeServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Internal server error"))
	w.WriteHeader(http.StatusInternalServerError)
}

func (p *JetStreamGetMsg) getStream(name string) (jetstream.Stream, error) {
	if p.stream != nil {
		return p.stream, nil
	}
	js, err := p.Client.JetStream()
	if err != nil {
		return nil, err
	}
	var streamName = p.Stream
	if streamName == "" {
		stream, err := js.StreamNameBySubject(p.ctx, p.Subject)
		if err != nil {
			return nil, err
		}
		streamName = stream
	}
	stream, err := js.Stream(p.ctx, streamName)
	if err != nil {
		return nil, err
	}
	p.stream = stream
	return p.stream, nil
}

// ServeHTTP receives incoming HTTP requests and is responsible
// for publishing request as a NATS message before writing to
// the HTTP response writer.
func (p JetStreamGetMsg) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	stream, err := p.getStream(p.Stream)
	if err != nil {
		p.logger.Error("Error getting NATS stream", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	var entry *jetstream.RawStreamMsg
	if p.Subject == "" && p.Sequence != -1 {
		entry, err = stream.GetMsg(p.ctx, uint64(p.Sequence))
	} else if p.Subject == "" && p.Sequence == -1 {
		entry, err = stream.GetLastMsgForSubject(p.ctx, ">")
	} else {
		entry, err = stream.GetLastMsgForSubject(p.ctx, p.Subject)
	}
	switch err {
	case jetstream.ErrMsgNotFound:
		w.WriteHeader(http.StatusNotFound)
		return nil
	case nil:
		break
	default:
		p.logger.Error("Error fetching message from stream", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	w.Header().Add("Nats-Js-Sequence", fmt.Sprintf("%d", entry.Sequence))
	w.Header().Add("Nats-Js-Time", fmt.Sprintf("%d", entry.Time.Unix()))
	for k, headers := range entry.Header {
		for _, v := range headers {
			w.Header().Add(fmt.Sprintf("Nats-Hdr-%s", k), v)
		}
	}
	w.Write(entry.Data)
	w.WriteHeader(http.StatusOK)
	p.logger.Warn("HTTP request processed", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	return next.ServeHTTP(w, r)
}
