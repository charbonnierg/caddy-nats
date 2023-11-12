package eventgridreceiver

import (
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventgrid"
	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"github.com/quara-dev/beyond/pkg/azutils"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AzureEventGridReader{})
}

func (AzureEventGridReader) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.readers.azure_eventgrid",
		New: func() caddy.Module { return new(AzureEventGridReader) },
	}
}

type AzureEventGridReader struct {
	ctx    caddy.Context
	logger *zap.Logger
	client *azeventgrid.Client

	Credentials      *azutils.CredentialConfig `json:"credentials,omitempty"`
	Endpoint         string                    `json:"endpoint,omitempty"`
	TopicName        string                    `json:"topic_name,omitempty"`
	SubscriptionName string                    `json:"subscription_name,omitempty"`
}

func (r *AzureEventGridReader) Open(ctx caddy.Context, _ *natsclient.NatsClient) error {
	r.ctx = ctx
	r.logger = ctx.Logger().Named("receiver.azure_eventgrid")
	r.logger.Info("provisioning azure eventgrid receiver", zap.String("endpoint", r.Endpoint), zap.String("topic_name", r.TopicName), zap.String("subscription_name", r.SubscriptionName))
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
	r.logger.Info("successfully provisioned azure eventgrid receiver", zap.String("endpoint", r.Endpoint), zap.String("topic_name", r.TopicName), zap.String("subscription_name", r.SubscriptionName))
	r.client = client
	return nil
}

func (r *AzureEventGridReader) Close() error {
	return nil
}

func (r *AzureEventGridReader) Read() (caddynats.Message, func() error, error) {
	var events *azeventgrid.ReceiveCloudEventsResponse
	for {
		response, err := r.client.ReceiveCloudEvents(r.ctx, r.TopicName, r.SubscriptionName, &azeventgrid.ReceiveCloudEventsOptions{
			// Receive only one event at a time.
			MaxEvents: to.Ptr[int32](1),
			// Wait for 30 seconds for events to arrive.
			MaxWaitTime: to.Ptr[int32](30),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to receive cloud events: %s", err.Error())
		}
		if len(response.Value) == 0 {
			continue
		}
		events = &response
		break
	}
	props := events.Value[0]
	cloudevent := props.Event
	r.logger.Info("received event", zap.Any("event", cloudevent))
	ack := func() error {
		// This acknowledges the event and causes it to be deleted from the subscription.
		// Other options are:
		// - client.ReleaseCloudEvents, which invalidates our event lock and allows another subscriber to receive the event.
		// - client.RejectCloudEvents, which rejects the event.
		//     If dead-lettering is configured, the event will be moved into the dead letter queue.
		//     Otherwise the event is deleted.
		ackResp, err := r.client.AcknowledgeCloudEvents(r.ctx, r.TopicName, r.SubscriptionName, azeventgrid.AcknowledgeOptions{
			LockTokens: []string{
				*props.BrokerProperties.LockToken,
			},
		}, nil)

		if err != nil {
			return fmt.Errorf("failed to acknowledge cloud events: %s", err.Error())
		}

		if len(ackResp.FailedLockTokens) > 0 {
			// some events failed when we tried to acknowledge them.
			for _, failed := range ackResp.FailedLockTokens {
				r.logger.Error("failed to acknowledge event", zap.String("lock_token", *failed.LockToken), zap.String("error_description", *failed.ErrorDescription))
			}
			return fmt.Errorf("failed to acknowledge cloud events: %s", err.Error())
		}
		return nil
	}

	return &cloudEventMsg{&props.Event}, ack, nil
}

type cloudEventMsg struct {
	msg *messaging.CloudEvent
}

func (m *cloudEventMsg) Payload() ([]byte, error) {
	return m.msg.MarshalJSON()
}

func (m *cloudEventMsg) Subject(prefix string) (string, error) {
	if m.msg.Subject != nil {
		sub := *m.msg.Subject
		if sub == "" {
			return "", errors.New("subject is empty")
		}
		if sub[0] == '/' {
			sub = sub[1:]
		}
		if sub == "" {
			return "", errors.New("subject is empty")
		}
		return prefix + "." + sub, nil
	}
	return "", errors.New("subject is empty")
}

func (m *cloudEventMsg) Headers() (map[string][]string, error) {
	return map[string][]string{
		"Event-Type": {m.msg.Type},
		"Event-Id":   {m.msg.ID},
		"Event-Time": {m.msg.Time.Format(time.RFC3339)},
	}, nil
}

var (
	_ caddynats.Reader = (*AzureEventGridReader)(nil)
)
