// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"go.uber.org/zap"
)

func NewLock(logger *zap.Logger, kvstore KeyValueStore, key string) sessions.Lock {
	key = "lock." + key
	return &Lock{logger: logger, kvstore: kvstore, key: key}
}

// Lock is a distributed lock implementation for JetStream.
type Lock struct {
	logger   *zap.Logger
	_lock    sync.Mutex
	kvstore  KeyValueStore
	key      string
	info     Expiration
	revision uint64
}

// Obtain obtains the lock on the distributed
// lock resource if no lock exists yet.
// Otherwise it will return ErrLockNotObtained
func (l *Lock) Obtain(ctx context.Context, expiration time.Duration) error {
	// NATS guarantees that a single caller will obtain the lock
	// because of the revision number.
	// But let's use a mutex to be sure and to avoid making
	// calls to server if we already have the lock.
	l._lock.Lock()
	defer l._lock.Unlock()
	// If the current lock is expired, try to obtain it
	if l.info.expired() {
		l.info.update(expiration)
		if err := l.updateRemote(); err != nil {
			l.logger.Error("failed to obtain lock", zap.Error(err))
			return sessions.ErrLockNotObtained
		}
		return nil
	}
	return sessions.ErrLockNotObtained
}

// Peek returns true if the lock currently exists
// Otherwise it returns false.
func (l *Lock) Peek(ctx context.Context) (bool, error) {
	last, info, err := l.getRemote()
	if err != nil {
		if err == nats.ErrKeyNotFound {
			l.logger.Error("missing lock", zap.Error(err))
			return false, err
		}
		l.logger.Error("failed to peek lock", zap.Error(err))
		return false, err
	}
	if last == l.revision && !info.expired() {
		return true, nil
	}
	return false, nil
}

// Refresh refreshes the expiration time of the lock,
// if is still applied.
// Otherwise it will return ErrNotLocked
func (l *Lock) Refresh(ctx context.Context, expiration time.Duration) error {
	// NATS guarantees that a single caller will release the lock
	// because of the revision number.
	// But let's use a mutex to to avoid making calls that will fail
	// to the server if we don't have the lock.
	l._lock.Lock()
	defer l._lock.Unlock()
	// If the current lock is expired, return an error
	if l.info.expired() {
		return sessions.ErrNotLocked
	}
	// Otherwise, update the expiration time
	l.info.update(expiration)
	if err := l.updateRemote(); err != nil {
		l.logger.Error("failed to refresh lock", zap.Error(err))
		return err
	}
	return nil
}

// Release removes the existing lock,
// Otherwise it will return ErrNotLocked
func (l *Lock) Release(ctx context.Context) error {
	// NATS guarantees that a single caller will release the lock
	// because of the revision number.
	// But let's use a mutex to to avoid making calls that will fail
	// to the server if we don't have the lock.
	l._lock.Lock()
	defer l._lock.Unlock()
	// If the current lock is expired, return an error
	if l.info.expired() {
		return sessions.ErrNotLocked
	}
	// Otherwise, remove the lock
	l.info.update(0)
	if err := l.updateRemote(); err != nil {
		l.logger.Error("failed to release lock", zap.Error(err))
		return sessions.ErrNotLocked
	}
	return nil
}

// getRemote can be called without holding the internal lock
func (l *Lock) getRemote() (uint64, *Expiration, error) {
	kv, err := l.kvstore.lazy()
	if err != nil {
		return 0, nil, err
	}
	item, err := kv.Get(l.kvstore.ctx, l.key)
	if err != nil {
		return 0, nil, err
	}
	revision := item.Revision()
	var info Expiration
	err = json.Unmarshal(item.Value(), &info)
	if err != nil {
		return revision, nil, err
	}
	return revision, &info, nil
}

// updateRemote must be called while holding the internal lock
func (l *Lock) updateRemote() error {
	kv, err := l.kvstore.lazy()
	if err != nil {
		return err
	}
	payload, err := l.info.encode()
	if err != nil {
		return err
	}
	newRevision, err := kv.Update(l.kvstore.ctx, l.key, payload, l.revision)
	if err != nil {
		return err
	}
	l.revision = newRevision
	return nil
}
