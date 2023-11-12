package natsclient

import (
	"context"
	"errors"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

type Stream struct {
	*jetstream.StreamConfig
	Prefix string `json:"prefix"`
}

func (d *Stream) Configure(ctx context.Context, js jetstream.JetStream) error {
	if d.StreamConfig == nil {
		return errors.New("stream config is nil")
	}
	if d.StreamConfig.Subjects == nil && d.Prefix != "" {
		d.StreamConfig.Subjects = []string{d.Prefix + ".>"}
	}
	_, err := js.CreateOrUpdateStream(ctx, *d.StreamConfig)
	if err != nil {
		return err
	}
	return nil
}

func (stream *Stream) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if stream.StreamConfig == nil {
		stream.StreamConfig = &jetstream.StreamConfig{}
	}
	if d.CountRemainingArgs() > 0 {
		if err := parser.ParseString(d, &stream.Name); err != nil {
			return err
		}
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "name":
			if err := parser.ParseString(d, &stream.Name); err != nil {
				return err
			}
		case "description":
			if err := parser.ParseString(d, &stream.Description); err != nil {
				return err
			}
		case "prefix":
			if err := parser.ParseString(d, &stream.Prefix); err != nil {
				return err
			}
		case "subjects":
			if err := parser.ParseStringArray(d, &stream.Subjects); err != nil {
				return err
			}
		case "retention":
			if err := parseRetentionPolicy(d, &stream.Retention); err != nil {
				return err
			}
		case "max_consumers":
			if err := parser.ParseInt(d, &stream.MaxConsumers); err != nil {
				return err
			}
		case "max_msgs":
			if err := parser.ParseInt64(d, &stream.MaxMsgs); err != nil {
				return err
			}
		case "max_bytes":
			if err := parser.ParseInt64ByteSize(d, &stream.MaxBytes); err != nil {
				return err
			}
		case "discard":
			if err := parseDiscardPolicy(d, &stream.Discard); err != nil {
				return err
			}
		case "discard_new_per_subject":
			if err := parser.ParseBool(d, &stream.DiscardNewPerSubject); err != nil {
				return err
			}
		case "max_age":
			if err := parser.ParseDuration(d, &stream.MaxAge); err != nil {
				return err
			}
		case "max_msgs_per_subject":
			if err := parser.ParseInt64(d, &stream.MaxMsgsPerSubject); err != nil {
				return err
			}
		case "max_msg_size":
			if err := parser.ParseInt32ByteSize(d, &stream.MaxMsgSize); err != nil {
				return err
			}
		case "storage":
			if err := parseStorage(d, &stream.Storage); err != nil {
				return err
			}
		case "replicas":
			if err := parser.ParseInt(d, &stream.Replicas); err != nil {
				return err
			}
		case "no_ack":
			if err := parser.ParseBool(d, &stream.NoAck); err != nil {
				return err
			}
		case "ack":
			if err := parser.ParseBool(d, &stream.NoAck, parser.Reverse()); err != nil {
				return err
			}
		case "template":
			if err := parser.ParseString(d, &stream.Template); err != nil {
				return err
			}
		case "duplicates":
			if err := parser.ParseDuration(d, &stream.Duplicates); err != nil {
				return err
			}
		case "cluster":
			stream.Placement = fnutils.DefaultIfNil(stream.Placement, &jetstream.Placement{})
			if err := parser.ParseString(d, &stream.Placement.Cluster); err != nil {
				return err
			}
		case "tag":
			stream.Placement = fnutils.DefaultIfNil(stream.Placement, &jetstream.Placement{})
			if err := parser.ParseStringArray(d, &stream.Placement.Tags); err != nil {
				return err
			}
		case "mirror":
			stream.Mirror = fnutils.DefaultIfNil(stream.Mirror, &jetstream.StreamSource{})
			if err := parseStreamSource(d, stream.Mirror); err != nil {
				return err
			}
		case "source":
			stream.Sources = fnutils.DefaultIfEmpty(stream.Sources, []*jetstream.StreamSource{})
			source := &jetstream.StreamSource{}
			if err := parseStreamSource(d, source); err != nil {
				return err
			}
			stream.Sources = append(stream.Sources, source)
		case "sealed":
			if err := parser.ParseBool(d, &stream.Sealed); err != nil {
				return err
			}
		case "deny_delete":
			if err := parser.ParseBool(d, &stream.DenyDelete); err != nil {
				return err
			}
		case "deny_purge":
			if err := parser.ParseBool(d, &stream.DenyPurge); err != nil {
				return err
			}
		case "allow_rollup":
			if err := parser.ParseBool(d, &stream.AllowRollup); err != nil {
				return err
			}
		case "first_seq":
			if err := parser.ParseUint64(d, &stream.FirstSeq); err != nil {
				return err
			}
		case "subject_transform":
			transform := jetstream.SubjectTransformConfig{}
			if err := parser.ParseString(d, &transform.Source); err != nil {
				return err
			}
			if err := parser.ExpectString(d, parser.Match("to")); err != nil {
				return err
			}
			if err := parser.ParseString(d, &transform.Destination); err != nil {
				return err
			}
			stream.SubjectTransform = &transform
		case "republish":
			stream.RePublish = fnutils.DefaultIfNil(stream.RePublish, &jetstream.RePublish{})
			if err := parseRePublish(d, stream.RePublish); err != nil {
				return err
			}
		case "allow_direct":
			if err := parser.ParseBool(d, &stream.AllowDirect); err != nil {
				return err
			}
		case "mirror_direct":
			if err := parser.ParseBool(d, &stream.MirrorDirect); err != nil {
				return err
			}
		case "consumer_inactive_threshold":
			if err := parser.ParseDuration(d, &stream.ConsumerLimits.InactiveThreshold); err != nil {
				return err
			}
		case "consumer_max_ack_pending":
			if err := parser.ParseInt(d, &stream.ConsumerLimits.MaxAckPending); err != nil {
				return err
			}
		case "metadata":
			if err := parser.ParseStringMap(d, &stream.Metadata); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}
	return nil
}
