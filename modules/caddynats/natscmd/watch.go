// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/nats-io/nats.go"
)

func watchCmd(fs caddycmd.Flags) (int, error) {
	clients, err := connect(fs)
	if err != nil {
		return 1, err
	}
	nc, _ := clients.Nats()
	defer nc.Close()
	js, _ := clients.JetStreamContext()
	stream := fs.Arg(0)
	filter, err := fs.GetStringArray("filter")
	if err != nil {
		return 1, err
	}
	info, err := js.AddConsumer(
		stream,
		&nats.ConsumerConfig{
			DeliverPolicy:  nats.DeliverNewPolicy,
			AckPolicy:      nats.AckNonePolicy,
			MaxWaiting:     1,
			MaxDeliver:     1,
			FilterSubjects: filter,
			ReplayPolicy:   nats.ReplayInstantPolicy,
		},
	)
	if err != nil {
		return 1, err
	}
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-shutdown
		cancel()
	}()
	sub, err := js.PullSubscribe("", "", nats.Bind(stream, info.Name), nats.Context(ctx))
	if err != nil {
		return 1, err
	}
	for {
		deadline, _cancel := context.WithTimeout(ctx, time.Duration(1)*time.Second)
		msgs, err := sub.Fetch(1, nats.Context(deadline))
		_cancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				continue
			}
			if err == context.Canceled {
				return 0, nil
			}
			return 1, err
		}
		for _, msg := range msgs {
			if msg == nil {
				return 0, nil
			}
			meta, _ := msg.Metadata()
			fmt.Printf("Received message on %s (stream_sequence=%d, consumer_sequence=%d):\n", msg.Subject, meta.Sequence.Stream, meta.Sequence.Consumer)
			fmt.Printf("%s\n", string(msg.Data))
		}
	}
}
