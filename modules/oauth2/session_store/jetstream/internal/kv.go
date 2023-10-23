// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// KeyValueStore is a JetStream key-value store.
// It is lazy and will only connect the first time
// kv method is called.
type KeyValueStore struct {
	natskv nats.KeyValue
	name   string
	ttl    time.Duration
	client *Client
}

func (s *KeyValueStore) kv() (nats.KeyValue, error) {
	if s.natskv == nil {
		js, err := s.client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jetstream: %v", err)
		}
		kv, err := js.KeyValue(s.name)
		if err != nil {
			if err == nats.ErrBucketNotFound {
				// Let's create the key value store
				kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
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
