package mongo_output

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/quara-dev/beyond/pkg/natsutils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Replicator pulls data from a JetStream stream and inserts it into MongoDB.
// A replicator may pull data for multiple collections, but it will only
// pull from one stream, and insert into one database.
type Replicator struct {
	ctx         context.Context
	db          *mongo.Database
	collections map[string]*mongo.Collection
	logger      *zap.Logger
	done        chan error
	conn        *natsutils.Connection
	stream      string
	consumer    string
	subjects    []string
}

func (r *Replicator) Start() error {
	r.collections = make(map[string]*mongo.Collection)
	go r.loop()
	return nil
}

func (r *Replicator) loop() {
	for {
		r.logger.Info("connecting to JetStream", zap.String("stream", r.stream))
		stream, err := r.conn.JetStream().StreamInfo(r.stream)
		if err != nil {
			r.logger.Error("error getting stream info", zap.Error(err))
			r.done <- err
			return
		}
		// Call to pull subscribe
		con, err := r.conn.JetStream().AddConsumer(stream.Config.Name, &nats.ConsumerConfig{
			Durable:        r.consumer,
			FilterSubjects: r.subjects,
			AckPolicy:      nats.AckExplicitPolicy,
			DeliverPolicy:  nats.DeliverAllPolicy,
			ReplayPolicy:   nats.ReplayInstantPolicy,
			MaxDeliver:     5,
			MaxWaiting:     1,
		})
		if err != nil {
			r.logger.Error("error adding consumer", zap.Error(err))
			continue
		}
		sub, err := r.conn.JetStream().PullSubscribe("", con.Config.Durable, nats.ConsumerFilterSubjects(r.subjects...))
		if err != nil {
			r.logger.Error("error subscribing to JetStream", zap.Error(err))
			continue
		}
		for {
			ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
			msg, err := sub.NextMsgWithContext(ctx)
			cancel()
			if err != nil {
				r.logger.Error("error reading message from JetStream", zap.Error(err))
				continue
			}
			// Insert into MongoDB
			err = r.processData(msg.Data)
			if err != nil {
				r.logger.Error("error processing data", zap.Error(err))
				continue
			}
		}
	}
}

func (r *Replicator) processData(data []byte) error {
	r.logger.Warn("processing data", zap.ByteString("data", data))
	return nil
}
