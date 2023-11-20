// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package eventgridwriter

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventgrid"
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/azutils"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AzureEventGridWriter{})
}

func (AzureEventGridWriter) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats.writers.azure_eventgrid",
		New: func() caddy.Module { return new(AzureEventGridWriter) },
	}
}

type AzureEventGridWriter struct {
	ctx    caddy.Context
	logger *zap.Logger
	client *azeventgrid.Client

	Credentials *azutils.CredentialConfig `json:"credentials,omitempty"`
	Endpoint    string                    `json:"endpoint,omitempty"`
	TopicName   string                    `json:"topic_name,omitempty"`
	EventType   string                    `json:"event_type,omitempty"`
}

func (r *AzureEventGridWriter) Open(ctx caddy.Context, _ *natsclient.NatsClient) error {
	r.ctx = ctx
	r.logger = ctx.Logger().Named("writer.azure_eventgrid")
	r.logger.Info("provisioning azure eventgrid writer", zap.String("endpoint", r.Endpoint), zap.String("topic_name", r.TopicName))
	if r.Credentials == nil {
		r.Credentials = &azutils.CredentialConfig{}
		r.Credentials.ParseEnv()
	}
	key, err := r.Credentials.AuthenticateWithAccessKey()
	if err != nil {
		return fmt.Errorf("failed to authenticate with access key: %s", err.Error())
	}
	client, err := azeventgrid.NewClientWithSharedKeyCredential(r.Endpoint, key, nil)
	if err != nil {
		return fmt.Errorf("failed to create eventgrid client: %s", err.Error())
	}
	r.logger.Info("successfully provisioned azure eventgrid writer", zap.String("endpoint", r.Endpoint), zap.String("topic_name", r.TopicName))
	r.client = client
	return nil
}

func (r *AzureEventGridWriter) Close() error {
	return nil
}

func (r *AzureEventGridWriter) Write(msg caddynats.Message) error {
	source, err := msg.Subject("")
	if err != nil {
		return err
	}
	if strings.HasPrefix(source, ".") {
		source = "/" + source[1:]
	} else if !strings.HasPrefix(source, "/") {
		source = "/" + source
	}
	payload, err := msg.Payload()
	if err != nil {
		return err
	}
	event, err := messaging.NewCloudEvent(source, r.EventType, payload, nil)

	if err != nil {
		return err
	}

	eventsToSend := []messaging.CloudEvent{
		event,
	}

	// NOTE: we're sending a single event as an example. For better efficiency it's best if you send
	// multiple events at a time.
	r.logger.Debug("sending cloud event", zap.Any("events", eventsToSend))
	_, err = r.client.PublishCloudEvents(r.ctx, r.TopicName, eventsToSend, nil)

	if err != nil {
		return err
	}
	r.logger.Info("successfully sent cloud event", zap.String("source", event.Source), zap.String("type", event.Type))
	return nil
}

var (
	_ caddynats.Writer = (*AzureEventGridWriter)(nil)
)
