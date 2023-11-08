package streamexporter

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(StreamExporter{})
}

type StreamExporter struct {
	ctx    caddy.Context
	js     nats.JetStreamContext
	logger *zap.Logger

	*client.Stream
}

func (StreamExporter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.exporters.stream",
		New: func() caddy.Module { return new(StreamExporter) },
	}
}

func (s *StreamExporter) Provision(ctx caddy.Context) error {
	s.ctx = ctx
	s.logger = ctx.Logger().Named("exporter.stream")
	s.logger.Info("provisioning NATS stream exporter", zap.String("name", s.Name), zap.String("prefix", s.Prefix))
	if s.StreamConfig.Subjects == nil && s.Prefix == "" {
		return errors.New("subjects and prefix are empty")
	}
	if s.StreamConfig.Subjects == nil {
		s.StreamConfig.Subjects = []string{s.Prefix + ".>"}
	}
	return nil
}

func (s *StreamExporter) Write(msg client.Message) error {
	sub, err := msg.Subject(s.Prefix)
	if err != nil {
		return err
	}
	payload, err := msg.Payload()
	if err != nil {
		return err
	}
	s.logger.Info("Publishing to NATS", zap.String("subject", sub), zap.ByteString("payload", payload))
	if _, err := s.js.Publish(sub, payload, nats.ExpectStream(s.Name)); err != nil {
		return err
	}
	return nil
}

func (s *StreamExporter) Connect(clients *client.Clients) error {
	if err := s.Stream.Configure(s.ctx, clients); err != nil {
		return err
	}
	s.js = clients.JetStream()
	return nil
}

func (s *StreamExporter) Close() error {
	return nil
}

var (
	_ client.Exporter = (*StreamExporter)(nil)
)
