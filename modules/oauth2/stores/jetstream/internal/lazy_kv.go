// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"go.uber.org/zap"
)

// KeyValueStore is a JetStream key-value store.
// It is lazy and will only connect the first time
// kv method is called.
type KeyValueStore struct {
	logger *zap.Logger
	natskv nats.KeyValue
	name   string
	ttl    time.Duration
	client *natsutils.Client
}

func (s *KeyValueStore) lazy() (nats.KeyValue, error) {
	if s.natskv == nil {
		s.logger.Info("connecting to jetstream")
		conn, err := s.client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jetstream: %v", err)
		}
		kv, err := conn.JetStream().KeyValue(s.name)
		if err != nil {
			if err == nats.ErrBucketNotFound {
				s.logger.Info("creating jetstream key-value store")
				// Let's create the key value store
				kv, err = conn.JetStream().CreateKeyValue(&nats.KeyValueConfig{
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
