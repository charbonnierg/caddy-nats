// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.uber.org/zap"
)

// KeyValueStore is a JetStream key-value store.
// It is lazy and will only connect the first time
// kv method is called.
type KeyValueStore struct {
	ctx    context.Context
	logger *zap.Logger
	natskv jetstream.KeyValue
	name   string
	ttl    time.Duration
	client *natsclient.NatsClient
}

func (s *KeyValueStore) lazy() (jetstream.KeyValue, error) {
	if s.natskv == nil {
		s.logger.Info("connecting to jetstream")
		js, err := s.client.JetStream()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jetstream: %v", err)
		}
		kv, err := js.KeyValue(s.ctx, s.name)
		if err != nil {
			if err == jetstream.ErrBucketNotFound {
				s.logger.Info("creating jetstream key-value store")
				// Let's create the key value store
				kv, err = js.CreateKeyValue(s.ctx, jetstream.KeyValueConfig{
					Bucket:      s.name,
					Description: "oauth2-proxy session store",
					TTL:         time.Duration(s.ttl),
					History:     1,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create key-value store: %v", err)
				}
			} else {
				return nil, fmt.Errorf("failed to lookup key-value store: %v", err)
			}
		}
		s.natskv = kv
	}
	return s.natskv, nil
}
