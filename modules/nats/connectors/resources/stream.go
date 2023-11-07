package resources

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
)

type Stream struct {
	*nats.StreamConfig
	Prefix string `json:"prefix"`
}

func (d *Stream) Configure(ctx context.Context, clients *Clients) error {
	if d.StreamConfig == nil {
		return errors.New("stream config is nil")
	}
	if d.StreamConfig.Subjects == nil {
		d.StreamConfig.Subjects = []string{d.Prefix + ".>"}
	}
	js := clients.JetStream()
	current, err := js.StreamInfo(d.Name, nats.Context(ctx))
	switch err {
	case nats.ErrStreamNotFound:
		_, err = js.AddStream(d.StreamConfig, nats.Context(ctx))
		if err != nil {
			return err
		}
	case nil:
		if d.isApproximatelyEqualTo(&current.Config) {
			return nil
		}
		_, err = js.UpdateStream(d.StreamConfig, nats.Context(ctx))
		if err != nil {
			return err
		}
	default:
		return err
	}
	return nil
}

func (d *Stream) isApproximatelyEqualTo(other *nats.StreamConfig) bool {
	if other == nil {
		return false
	}
	if d.Name != other.Name {
		return false
	}
	if d.Subjects != nil && other.Subjects != nil {
		if len(d.Subjects) != len(other.Subjects) {
			return false
		}
		for i, subject := range d.Subjects {
			if subject != other.Subjects[i] {
				return false
			}
		}
	}
	if d.Retention != other.Retention {
		return false
	}
	if d.MaxConsumers != other.MaxConsumers {
		return false
	}
	if d.MaxMsgs != other.MaxMsgs {
		return false
	}
	if d.MaxBytes != other.MaxBytes {
		return false
	}
	if d.MaxAge != other.MaxAge {
		return false
	}
	if d.MaxMsgSize != other.MaxMsgSize {
		return false
	}
	if d.Discard != other.Discard {
		return false
	}
	if d.DiscardNewPerSubject != other.DiscardNewPerSubject {
		return false
	}
	if d.NoAck != other.NoAck {
		return false
	}
	return true
}

var (
	_ resource = (*Stream)(nil)
)
