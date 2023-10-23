// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"github.com/nats-io/nats-server/v2/server"
)

type NatsApp interface {
	ServerProvider
}

type ServerProvider interface {
	GetServer() (*server.Server, error)
}
