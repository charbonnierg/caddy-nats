package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func ParseKeyValueStore(d *caddyfile.Dispenser, c *client.Connection) error {
	store := &client.KeyValueStore{
		KeyValueConfig: &nats.KeyValueConfig{},
	}
	if err := parser.ParseString(d, &store.Bucket); err != nil {
		return err
	}
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "description":
			if err := parser.ParseString(d, &store.Description); err != nil {
				return err
			}
		case "max_value_size":
			if err := parser.ParseInt32ByteSize(d, &store.MaxValueSize); err != nil {
				return err
			}
		case "history":
			if err := parser.ParseUint8(d, &store.History); err != nil {
				return err
			}
		case "ttl":
			if err := parser.ParseDuration(d, &store.TTL); err != nil {
				return err
			}
		case "max_bytes":
			if err := parser.ParseInt64ByteSize(d, &store.MaxBytes); err != nil {
				return err
			}
		case "storage":
			if err := parseStorage(d, &store.Storage); err != nil {
				return err
			}
		case "replicas":
			if err := parser.ParseInt(d, &store.Replicas); err != nil {
				return err
			}
		case "republish":
			store.RePublish = fnutils.DefaultIfNil(store.RePublish, &nats.RePublish{})
			if err := parseRePublish(d, store.RePublish); err != nil {
				return err
			}
		case "mirror":
			store.Mirror = fnutils.DefaultIfNil(store.Mirror, &nats.StreamSource{})
			if err := parseStreamSource(d, store.Mirror); err != nil {
				return err
			}
		case "source":
			store.Sources = fnutils.DefaultIfEmpty(store.Sources, []*nats.StreamSource{})
			source := &nats.StreamSource{}
			if err := parseStreamSource(d, source); err != nil {
				return err
			}
			store.Sources = append(store.Sources, source)
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}

func parseRePublish(d *caddyfile.Dispenser, dest *nats.RePublish) error {
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

func parseStreamSource(d *caddyfile.Dispenser, dest *nats.StreamSource) error {
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
			dest.SubjectTransforms = fnutils.DefaultIfEmpty(dest.SubjectTransforms, []nats.SubjectTransformConfig{})
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
			dest.SubjectTransforms = append(dest.SubjectTransforms, transform)
		case "api_prefix":
			dest.External = fnutils.DefaultIfNil(dest.External, &nats.ExternalStream{})
			if err := parser.ParseString(d, &dest.External.APIPrefix); err != nil {
				return err
			}
		case "deliver_prefix":
			dest.External = fnutils.DefaultIfNil(dest.External, &nats.ExternalStream{})
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
