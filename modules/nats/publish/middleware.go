package publish

import (
	"io"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	nc "github.com/nats-io/nats.go"
	caddynats "github.com/quara-dev/beyond/modules/nats"
	"github.com/quara-dev/beyond/modules/nats/client"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(NatsPublish{})
	httpcaddyfile.RegisterHandlerDirective("nats_publish", parseHandlerDirective)
}

func (NatsPublish) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.nats_publish",
		New: func() caddy.Module { return new(NatsPublish) },
	}
}

// NatsPublish is an http middleware which publishes incoming HTTP
// requests as NATS messages before replying with a 204 empty
// response.
type NatsPublish struct {
	logger     *zap.Logger
	Connection client.Connection `json:"connection,omitempty"`
	Subject    string            `json:"subject,omitempty"`
}

func (p *NatsPublish) Provision(ctx caddy.Context) error {
	app, err := caddynats.Load(ctx)
	if err != nil {
		return err
	}
	p.logger = ctx.Logger()
	if err := p.Connection.Provision(app); err != nil {
		return err
	}
	return nil
}

func (p NatsPublish) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	p.logger.Info("Publishing message to NATS", zap.String("subject", p.Subject))

	conn, err := p.Connection.Conn()

	if err != nil {
		p.logger.Error("Error getting NATS connection", zap.Error(err))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Error publishing message to NATS"))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	headers := nc.Header{}
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
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Error publishing message to NATS"))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	err = conn.PublishMsg(&nc.Msg{
		Subject: p.Subject,
		Data:    content,
		Header:  headers,
	})

	if err != nil {
		p.logger.Error("Error publishing message", zap.Error(err))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Error publishing message to NATS"))
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return next.ServeHTTP(w, r)
}
