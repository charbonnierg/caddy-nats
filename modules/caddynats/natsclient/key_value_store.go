// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsclient

import (
	"context"
	"errors"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

// KeyValueStore is a JetStream key value store. Key value stores are used to put/get arbitrary bytes
// values under string keys. It's also possible to delete keys, list keys, and watch changes for the
// whole store or a subset of keys.
type KeyValueStore struct {
	*jetstream.KeyValueConfig
}

// Configure creates or updates the key value store.
func (s *KeyValueStore) Configure(ctx context.Context, js jetstream.JetStream) error {
	if s.KeyValueConfig == nil {
		return errors.New("key value config is nil")
	}
	current, err := js.KeyValue(ctx, s.Bucket)
	switch err {
	case jetstream.ErrBucketNotFound:
		_, err = js.CreateKeyValue(ctx, *s.KeyValueConfig)
		if err != nil {
			return err
		}
	case nil:
		info, err := current.Status(ctx)
		if err != nil {
			return err
		}
		if s.isApproximatelyEqualTo(info) {
			return errors.New("key value store already exists with different config")
		}
		return nil
	default:
		return err
	}
	return nil
}

func (s *KeyValueStore) isApproximatelyEqualTo(other jetstream.KeyValueStatus) bool {
	if other == nil {
		return false
	}
	if s.Bucket != other.Bucket() {
		return false
	}
	if int64(s.History) != other.History() {
		return false
	}
	if s.TTL != other.TTL() {
		return false
	}
	if s.Storage.String() != other.BackingStore() {
		return false
	}
	return true
}

func (store *KeyValueStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if store.KeyValueConfig == nil {
		store.KeyValueConfig = &jetstream.KeyValueConfig{}
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
			store.RePublish = fnutils.DefaultIfNil(store.RePublish, &jetstream.RePublish{})
			if err := parseRePublish(d, store.RePublish); err != nil {
				return err
			}
		case "mirror":
			store.Mirror = fnutils.DefaultIfNil(store.Mirror, &jetstream.StreamSource{})
			if err := parseStreamSource(d, store.Mirror); err != nil {
				return err
			}
		case "source":
			store.Sources = fnutils.DefaultIfEmpty(store.Sources, []*jetstream.StreamSource{})
			source := &jetstream.StreamSource{}
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
