package caddyfile

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/modules/nats/client"
	"github.com/quara-dev/beyond/pkg/caddyutils/parser"
	"github.com/quara-dev/beyond/pkg/fnutils"
)

func ParseObjectStore(d *caddyfile.Dispenser, c *client.Connection) error {
	store := &client.ObjectStore{
		ObjectStoreConfig: &nats.ObjectStoreConfig{},
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
			if err := parseStorage(d, &store.Storage); err != nil {
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

func parseStorage(d *caddyfile.Dispenser, dest *nats.StorageType) error {
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
