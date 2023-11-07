package resources

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
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
func (s *ObjectStore) Configure(ctx context.Context, clients *Clients) error {
	if s.ObjectStoreConfig == nil {
		return errors.New("object store config is nil")
	}
	js := clients.JetStream()
	current, err := js.ObjectStore(s.Bucket)
	switch err {
	case nats.ErrBucketNotFound:
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
			return errors.New("object store already exists with different config")
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

var (
	_ resource = (*ObjectStore)(nil)
)
