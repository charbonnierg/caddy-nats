// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package jetstream

import (
	"encoding/json"
	"time"
)

type Expiration struct {
	Deadline time.Time
}

func (e *Expiration) update(expiration time.Duration) {
	e.Deadline = time.Now().Add(expiration)
}

func (e *Expiration) expired() bool {
	return time.Now().After(e.Deadline)
}

func (e *Expiration) encode() ([]byte, error) {
	return json.Marshal(e)
}

// func (e *Expiration) encodeWithPrefix() ([]byte, error) {
// 	payload, err := e.encode()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to encode expiration: %v", err)
// 	}
// 	prefix := []byte("expires=")
// 	suffix := []byte(";")
// 	encoded := append(
// 		append(prefix, payload[:]...),
// 		suffix[:]...,
// 	)
// 	return encoded, nil
// }

// func (e *Expiration) decode(payload []byte) error {
// 	return json.Unmarshal(payload, e)
// }

// func (e *Expiration) decodeWithPrefix(payload []byte) error {
// 	parts := strings.SplitN(string(payload), "=", 2)
// 	if len(parts) != 2 {
// 		return errors.New("invalid expiration format")
// 	}
// 	if parts[0] != "expires" {
// 		return errors.New("invalid expiration prefix")
// 	}
// 	err := e.decode([]byte(parts[1]))
// 	if err != nil {
// 		return errors.New("invalid expiration payload")
// 	}
// 	return nil
// }
