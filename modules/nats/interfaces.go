// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond"
)

type NatsApp interface {
	beyond.App
	Reload() error
	GetServer() (*server.Server, error)
	CreateClient(options ...nats.Option) (*nats.Conn, error)
	LoadBeyondApp(id string) (beyond.App, error)
}
