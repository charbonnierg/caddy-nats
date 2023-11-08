package client

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
)

// KeyValueStore is a JetStream key value store. Key value stores are used to put/get arbitrary bytes
// values under string keys. It's also possible to delete keys, list keys, and watch changes for the
// whole store or a subset of keys.
type KeyValueStore struct {
	*nats.KeyValueConfig
}

// Configure creates or updates the key value store.
func (s *KeyValueStore) Configure(ctx context.Context, clients *Clients) error {
	if s.KeyValueConfig == nil {
		return errors.New("key value config is nil")
	}
	js := clients.JetStream()
	current, err := js.KeyValue(s.Bucket)
	switch err {
	case nats.ErrBucketNotFound:
		_, err = js.CreateKeyValue(s.KeyValueConfig)
		if err != nil {
			return err
		}
	case nil:
		info, err := current.Status()
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

func (s *KeyValueStore) isApproximatelyEqualTo(other nats.KeyValueStatus) bool {
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

var (
	_ resource = (*KeyValueStore)(nil)
)
