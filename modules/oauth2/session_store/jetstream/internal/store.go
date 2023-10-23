// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"time"

	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/sessions/persistence"
	"go.uber.org/zap"
)

func NewStore(name string, client *Client, ttl time.Duration, logger *zap.Logger) *Store {
	return &Store{kvstore: KeyValueStore{name: name, client: client, ttl: ttl}, logger: logger}
}

type Store struct {
	kvstore KeyValueStore
	logger  *zap.Logger
}

func (s *Store) SessionStore(cookieOpts *options.Cookie) (sessions.SessionStore, error) {
	return persistence.NewManager(s, cookieOpts), nil
}

func (s *Store) Lock(key string) sessions.Lock {
	return NewLock(s.logger.Named("lock."+key), s.kvstore, key)
}

func (s *Store) Clear(ctx context.Context, key string) error {
	kv, err := s.kvstore.kv()
	if err != nil {
		s.logger.Error("failed to get kv", zap.Error(err))
		return err
	}
	return kv.Delete(key)
}

func (s *Store) Load(ctx context.Context, key string) ([]byte, error) {
	kv, err := s.kvstore.kv()
	if err != nil {
		s.logger.Error("failed to get kv", zap.Error(err))
		return nil, err
	}
	item, err := kv.Get(key)
	if err != nil {
		s.logger.Error("failed to get key", zap.Error(err))
		return nil, err
	}
	return decode(item.Value())
}

func (s *Store) VerifyConnection(ctx context.Context) error {
	kv, err := s.kvstore.kv()
	if err != nil {
		s.logger.Error("failed to get kv", zap.Error(err))
		return err
	}
	_, err = kv.Status()
	if err != nil {
		s.logger.Error("failed to get status", zap.Error(err))
		return err
	}
	return nil
}

func (s *Store) Save(ctx context.Context, key string, value []byte, expires time.Duration) error {
	kv, err := s.kvstore.kv()
	if err != nil {
		s.logger.Error("failed to get kv", zap.Error(err))
		return err
	}
	encoded, err := encode(value, expires)
	if err != nil {
		s.logger.Error("failed to encode value", zap.Error(err))
		return err
	}
	_, err = kv.Put(key, encoded)
	return err
}

// encode will encode arbitrary data with an expiration time.
func encode(value []byte, expires time.Duration) ([]byte, error) {
	// info := Expiration{}
	// info.update(expires)
	// prefix, err := info.encodeWithPrefix()
	// if err != nil {
	// 	return nil, err
	// }
	// return append(prefix[:], value[:]...), nil
	return value, nil
}

// decode will decode arbitrary data with an expiration time.
// if the data is expired, an error will be returned.
func decode(value []byte) ([]byte, error) {
	// parts := bytes.Split(value, []byte(";"))
	// if len(parts) != 2 {
	// 	return nil, errors.New("invalid format")
	// }
	// // Get expiration and actual data
	// expiration := parts[0]
	// data := parts[1]
	// // Decode expiration
	// info := Expiration{}
	// err := info.decodeWithPrefix(expiration)
	// if err != nil {
	// 	return nil, err
	// }
	// // Check expiration
	// if info.expired() {
	// 	return nil, errors.New("expired")
	// }
	// Return data
	// return data, nil
	return value, nil
}
