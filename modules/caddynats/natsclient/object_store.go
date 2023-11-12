// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package natsclient

import (
	"context"
	"errors"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

// ObjectStore is a JetStream object store. Object stores are used to put/get arbitrary bytes
// values under string keys. It's also possible to delete keys, list keys, and watch changes for the
// whole store or a subset of keys. Object stores are different from key value stores in that they
// support big objects (objects will be chunked and stored in the backing store) at the cost of higher
// latency and lower throughput.
type ObjectStore struct {
	*nats.ObjectStoreConfig
}

// Configure creates or updates the object store.
func (s *ObjectStore) Configure(ctx context.Context, js nats.JetStreamContext) error {
	if s.ObjectStoreConfig == nil {
		return errors.New("object store config is nil")
	}
	current, err := js.ObjectStore(s.Bucket)
	switch err {
	case nats.ErrStreamNotFound:
		_, err = js.CreateObjectStore(s.ObjectStoreConfig)
		if err != nil {
			return err
		}
	case nil:
		info, err := current.Status()
		if err != nil {
			return err
		}
		if s.isApproximatelyEqualTo(info) {
			return nil
		}
		return nil
	default:
		return err
	}
	return nil
}

func (s *ObjectStore) isApproximatelyEqualTo(other nats.ObjectStoreStatus) bool {
	if other == nil {
		return false
	}
	if s.Bucket != other.Bucket() {
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

func (store *ObjectStore) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if store.ObjectStoreConfig == nil {
		store.ObjectStoreConfig = &nats.ObjectStoreConfig{}
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
		case "ttl":
			if err := parser.ParseDuration(d, &store.TTL); err != nil {
				return err
			}
		case "max_bytes":
			if err := parser.ParseInt64ByteSize(d, &store.MaxBytes); err != nil {
				return err
			}
		case "storage":
			if err := parseStorageLegacy(d, &store.Storage); err != nil {
				return err
			}
		case "replicas":
			if err := parser.ParseInt(d, &store.Replicas); err != nil {
				return err
			}
		case "cluster":
			store.Placement = fnutils.DefaultIfNil(store.Placement, &nats.Placement{})
			if err := parser.ParseString(d, &store.Placement.Cluster); err != nil {
				return err
			}
		case "tag":
			store.Placement = fnutils.DefaultIfNil(store.Placement, &nats.Placement{})
			if err := parser.ParseStringArray(d, &store.Placement.Tags); err != nil {
				return err
			}
		case "metadata":
			if err := parser.ParseStringMap(d, &store.Metadata); err != nil {
				return err
			}
		default:
			return d.Errf("unrecognized subdirective: %s", d.Val())
		}
	}
	return nil
}
