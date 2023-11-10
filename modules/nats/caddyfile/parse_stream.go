package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func ParseStream(d *caddyfile.Dispenser, stream *client.Stream) error {
	if stream.StreamConfig == nil {
		stream.StreamConfig = &nats.StreamConfig{}
	}
	if err := parser.ParseString(d, &stream.Name); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
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
			if err := parseRetention(d, &stream.Retention); err != nil {
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
			if err := parseDiscard(d, &stream.Discard); err != nil {
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
			stream.Placement = fnutils.DefaultIfNil(stream.Placement, &nats.Placement{})
			if err := parser.ParseString(d, &stream.Placement.Cluster); err != nil {
				return err
			}
		case "tag":
			stream.Placement = fnutils.DefaultIfNil(stream.Placement, &nats.Placement{})
			if err := parser.ParseStringArray(d, &stream.Placement.Tags); err != nil {
				return err
			}
		case "mirror":
			stream.Mirror = fnutils.DefaultIfNil(stream.Mirror, &nats.StreamSource{})
			if err := parseStreamSource(d, stream.Mirror); err != nil {
				return err
			}
		case "source":
			stream.Sources = fnutils.DefaultIfEmpty(stream.Sources, []*nats.StreamSource{})
			source := &nats.StreamSource{}
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
			transform := nats.SubjectTransformConfig{}
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
			stream.RePublish = fnutils.DefaultIfNil(stream.RePublish, &nats.RePublish{})
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

func parseDiscard(d *caddyfile.Dispenser, policy *nats.DiscardPolicy) error {
	var discard string
	if err := parser.ParseString(d, &discard); err != nil {
		return err
	}
	switch discard {
	case "old":
		*policy = nats.DiscardOld
	case "new":
		*policy = nats.DiscardNew
	default:
		return d.Errf("unrecognized discard policy '%s'", discard)
	}
	return nil
}

func parseRetention(d *caddyfile.Dispenser, retention *nats.RetentionPolicy) error {
	var ret string
	if err := parser.ParseString(d, &ret); err != nil {
		return err
	}
	switch ret {
	case "limits":
		*retention = nats.LimitsPolicy
	case "interest":
		*retention = nats.InterestPolicy
	case "workqueue", "work_queue":
		*retention = nats.WorkQueuePolicy
	default:
		return d.Errf("unrecognized retention policy '%s'", ret)
	}
	return nil
}
