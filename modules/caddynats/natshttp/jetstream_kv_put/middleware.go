package jetstream_kv_put

import (
	"fmt"
	"io"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(JetStreamKeyValuePut{})
	httpcaddyfile.RegisterHandlerDirective("jetstream_put", parseHandlerDirective)
}

// CaddyModule implements the caddy.Module interface.
// It returns information about the caddy module.
func (JetStreamKeyValuePut) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.jetstream_put",
		New: func() caddy.Module { return new(JetStreamKeyValuePut) },
	}
}

// JetStreamKeyValuePut is an http middleware which publishes incoming HTTP
// requests as NATS messages and wait for an acknowledgement to be
// sent by JetStream engine (notifying that message has been persisted into a stream)
// before replying with a 204 empty response.
// HTTP headers are included in the response indicating
// which stream the message was persisted into, as well as
// the sequence the message was inserted to.
type JetStreamKeyValuePut struct {
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
func (p *JetStreamKeyValuePut) Provision(ctx caddy.Context) error {
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
func (p JetStreamKeyValuePut) writeServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Internal server error"))
	w.WriteHeader(http.StatusInternalServerError)
}

func (p *JetStreamKeyValuePut) getBucket(name string) (jetstream.KeyValue, error) {
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
func (p JetStreamKeyValuePut) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	bucket, err := p.getBucket(p.Bucket)
	if err != nil {
		p.logger.Error("Error getting NATS connection", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	headers := nats.Header{}
	for k, v := range r.Header {
		values, ok := headers[k]
		if !ok {
			values = []string{}
		}
		values = append(values, v...)
		headers[k] = values
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		p.logger.Error("Error reading request body", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	revision, err := bucket.Put(p.ctx, p.Key, content)
	if err != nil {
		p.logger.Error("Error publishing message to stream", zap.Error(err))
		p.writeServerError(w)
		return err
	}
	w.Header().Add("Nats-Js-Revision", fmt.Sprintf("%d", revision))
	w.WriteHeader(http.StatusNoContent)
	p.logger.Warn("HTTP request processed", zap.String("method", r.Method), zap.String("path", r.URL.Path))
	return next.ServeHTTP(w, r)
}
