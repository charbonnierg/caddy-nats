// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"time"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/nats-io/nats.go"
)

func pubCmd(fs caddycmd.Flags) (int, error) {
	clients, err := connect(fs)
	if err != nil {
		return 1, err
	}
	nc, _ := clients.Nats()
	defer nc.Close()
	subject := fs.Arg(0)
	sleep := fs.Duration("sleep")
	reply := fs.String("reply")
	count := parseCount(fs)
	payload, err := parsePayload(fs, 1)
	if err != nil {
		return 1, err
	}
	header, err := parseHeaders(fs)
	if err != nil {
		return 1, err
	}
	// Build NATS message
	msg := nats.Msg{
		Subject: subject,
		Data:    []byte(payload),
		Reply:   reply,
		Header:  header,
	}
	// Loop until count is reached
	for i := 0; i < count; i++ {
		if err := nc.PublishMsg(&msg); err != nil {
			return 1, fmt.Errorf("failed to publish on %s: %s", subject, err.Error())
		}
		// Sleep a bit if asked to
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
	return 0, nil
}
