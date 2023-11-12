package cmd

import (
	"fmt"
	"os"
	"os/signal"

	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/nats-io/nats.go"
)

func subCmd(fs caddycmd.Flags) (int, error) {
	clients, err := connect(fs)
	if err != nil {
		return 1, err
	}
	nc, _ := clients.Nats()
	defer nc.Close()
	subject := fs.Arg(0)
	queue := fs.String("queue")
	feed := make(chan *nats.Msg)
	count := 0
	sub, err := nc.QueueSubscribeSyncWithChan(subject, queue, feed)
	if err != nil {
		return 1, err
	}
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	for {
		select {
		case msg := <-feed:
			if msg == nil {
				return 0, nil
			}
			count += 1
			fmt.Printf("Received message on %s (count=%d):\n", msg.Subject, count)
			fmt.Printf("%s\n", string(msg.Data))
		case <-shutdown:
			sub.Unsubscribe()
			return 0, nil
		}
	}
}
