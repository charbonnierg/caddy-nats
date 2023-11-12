package natsclient

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func parseDiscardPolicy(d *caddyfile.Dispenser, policy *jetstream.DiscardPolicy) error {
	var discard string
	if err := parser.ParseString(d, &discard); err != nil {
		return err
	}
	switch discard {
	case "old":
		*policy = jetstream.DiscardOld
	case "new":
		*policy = jetstream.DiscardNew
	default:
		return d.Errf("unrecognized discard policy '%s'", discard)
	}
	return nil
}

func parseRetentionPolicy(d *caddyfile.Dispenser, retention *jetstream.RetentionPolicy) error {
	var ret string
	if err := parser.ParseString(d, &ret); err != nil {
		return err
	}
	switch ret {
	case "limits":
		*retention = jetstream.LimitsPolicy
	case "interest":
		*retention = jetstream.InterestPolicy
	case "workqueue", "work_queue":
		*retention = jetstream.WorkQueuePolicy
	default:
		return d.Errf("unrecognized retention policy '%s'", ret)
	}
	return nil
}

func parseDeliverPolicy(d *caddyfile.Dispenser, dest *jetstream.DeliverPolicy) error {
	var pol string
	if err := parser.ParseString(d, &pol); err != nil {
		return err
	}
	switch pol {
	case "last":
		*dest = jetstream.DeliverLastPolicy
	case "last_per_subject":
		*dest = jetstream.DeliverLastPerSubjectPolicy
	case "new":
		*dest = jetstream.DeliverNewPolicy
	case "start_sequence":
		*dest = jetstream.DeliverByStartSequencePolicy
	case "start_time":
		*dest = jetstream.DeliverByStartTimePolicy
	default:
		return d.Errf("unrecognized deliver policy '%s'", pol)
	}
	return nil
}

func parseAckPolicy(d *caddyfile.Dispenser, dest *jetstream.AckPolicy) error {
	var ack string
	if err := parser.ParseString(d, &ack); err != nil {
		return err
	}
	switch ack {
	case "all":
		*dest = jetstream.AckAllPolicy
	case "explicit":
		*dest = jetstream.AckExplicitPolicy
	case "none":
		*dest = jetstream.AckNonePolicy
	default:
		return d.Errf("unrecognized ack policy '%s'", ack)
	}
	return nil
}

func parseReplayPolicy(d *caddyfile.Dispenser, dest *jetstream.ReplayPolicy) error {
	var replay string
	if err := parser.ParseString(d, &replay); err != nil {
		return err
	}
	switch replay {
	case "instant":
		*dest = jetstream.ReplayInstantPolicy
	case "original":
		*dest = jetstream.ReplayOriginalPolicy
	default:
		return d.Errf("unrecognized replay policy '%s'", replay)
	}
	return nil
}

func parseStorage(d *caddyfile.Dispenser, dest *jetstream.StorageType) error {
	var storage string
	if err := parser.ParseString(d, &storage); err != nil {
		return err
	}
	switch storage {
	case "file":
		*dest = jetstream.FileStorage
	case "memory":
		*dest = jetstream.MemoryStorage
	default:
		return d.Errf("unrecognized storage '%s'", storage)
	}
	return nil
}

func parseStorageLegacy(d *caddyfile.Dispenser, dest *nats.StorageType) error {
	var storage string
	if err := parser.ParseString(d, &storage); err != nil {
		return err
	}
	switch storage {
	case "file":
		*dest = nats.FileStorage
	case "memory":
		*dest = nats.MemoryStorage
	default:
		return d.Errf("unrecognized storage '%s'", storage)
	}
	return nil
}

func parseRePublish(d *caddyfile.Dispenser, dest *jetstream.RePublish) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "source":
			if err := parser.ParseString(d, &dest.Source); err != nil {
				return err
			}
		case "destination":
			if err := parser.ParseString(d, &dest.Destination); err != nil {
				return err
			}
		case "headers_only":
			if err := parser.ParseBool(d, &dest.HeadersOnly); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}

func parseStreamSource(d *caddyfile.Dispenser, dest *jetstream.StreamSource) error {
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "name":
			if err := parser.ParseString(d, &dest.Name); err != nil {
				return err
			}
		case "opt_start_seq":
			if err := parser.ParseUint64(d, &dest.OptStartSeq); err != nil {
				return err
			}
		case "opt_start_time":
			if err := parser.ParseTime(d, dest.OptStartTime); err != nil {
				return err
			}
		case "filter_subject":
			if err := parser.ParseString(d, &dest.FilterSubject); err != nil {
				return err
			}
		case "subject_transform":
			dest.SubjectTransforms = fnutils.DefaultIfEmpty(dest.SubjectTransforms, []jetstream.SubjectTransformConfig{})
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
			dest.SubjectTransforms = append(dest.SubjectTransforms, transform)
		case "api_prefix":
			dest.External = fnutils.DefaultIfNil(dest.External, &jetstream.ExternalStream{})
			if err := parser.ParseString(d, &dest.External.APIPrefix); err != nil {
				return err
			}
		case "deliver_prefix":
			dest.External = fnutils.DefaultIfNil(dest.External, &jetstream.ExternalStream{})
			if err := parser.ParseString(d, &dest.External.DeliverPrefix); err != nil {
				return err
			}
		case "domain":
			if err := parser.ParseString(d, &dest.Domain); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
