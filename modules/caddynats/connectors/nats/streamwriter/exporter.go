package streamexporter

import (
	"errors"

	"github.com/caddyserver/caddy/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(StreamExporter{})
}

type StreamExporter struct {
	ctx    caddy.Context
	logger *zap.Logger
	js     jetstream.JetStream
	*natsclient.Stream
}

func (StreamExporter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.writers.stream",
		New: func() caddy.Module { return new(StreamExporter) },
	}
}

func (s *StreamExporter) Write(msg caddynats.Message) error {
	sub, err := msg.Subject(s.Prefix)
	if err != nil {
		return err
	}
	payload, err := msg.Payload()
	if err != nil {
		return err
	}
	s.logger.Info("Publishing to NATS", zap.String("subject", sub), zap.ByteString("payload", payload))
	if _, err := s.js.PublishMsg(s.ctx, &nats.Msg{Subject: sub, Data: payload}, jetstream.WithExpectStream(s.Name)); err != nil {
		return err
	}
	return nil
}

func (s *StreamExporter) Open(ctx caddy.Context, client *natsclient.NatsClient) error {
	s.logger.Info("opening NATS stream writer", zap.String("stream", s.Name), zap.Strings("subjects", s.Subjects))
	s.ctx = ctx
	s.logger = ctx.Logger().Named("writer.stream")
	if s.StreamConfig.Subjects == nil && s.Prefix == "" {
		return errors.New("subjects and prefix are empty")
	}
	if s.StreamConfig.Subjects == nil {
		s.StreamConfig.Subjects = []string{s.Prefix + ".>"}
	}
	if err := client.ConfigureStream(ctx, s.Stream); err != nil {
		return err
	}
	js, err := client.JetStream()
	if err != nil {
		return err
	}
	s.js = js
	return nil
}

func (s *StreamExporter) Close() error {
	return nil
}

var (
	_ caddynats.Writer = (*StreamExporter)(nil)
)
