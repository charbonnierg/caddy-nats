// Copyright 2023 QUARA - RGPI
// SPDX-License-Identifier: Apache-2.0

package changestreamreader

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/quara-dev/beyond/modules/caddynats"
	"github.com/quara-dev/beyond/modules/caddynats/connectors/mongo/common"
	"github.com/quara-dev/beyond/modules/caddynats/natsclient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ChangeStreamReader{})
}

func (ChangeStreamReader) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "nats_server.readers.mongodb_change_stream",
		New: func() caddy.Module { return new(ChangeStreamReader) },
	}
}

// ChangeStreamReader can be used to connect to a database
// and read a change stream.
// Receivers always have an associated stream where events
// are forwarded to.
type ChangeStreamReader struct {
	ctx                   caddy.Context
	client                *mongo.Client
	database              *mongo.Database
	collection            *mongo.Collection
	stream                *mongo.ChangeStream
	resumeTokenDatabase   *mongo.Database
	resumeTokenCollection *mongo.Collection
	logger                *zap.Logger

	Uri                   string `json:"uri"`
	Database              string `json:"database"`
	Collection            string `json:"collection"`
	ResumeTokenDatabase   string `json:"resume_token_database,omitempty"`
	ResumeTokenCollection string `json:"resume_token_collection,omitempty"`
}

// Connect to the database
func (r *ChangeStreamReader) Open(ctx caddy.Context, clients *natsclient.NatsClient) error {
	r.ctx = ctx
	r.logger = ctx.Logger().Named("receiver.mongodb_change_stream")
	r.logger.Info("provisioning mongodb change stream receiver", zap.String("uri", r.Uri))
	parsedUri, err := url.Parse(r.Uri)
	if err != nil {
		return fmt.Errorf("invalid mongodb uri: %v", err)
	}
	client, err := mongo.Connect(r.ctx, options.Client().ApplyURI(r.Uri))
	if err != nil {
		return fmt.Errorf("could not connect to mongodb: %v", err)
	}
	r.client = client
	r.logger.Info("connecting to mongodb", zap.String("uri", parsedUri.Redacted()))
	// Set database and collection
	if r.ResumeTokenDatabase == "" {
		r.ResumeTokenDatabase = r.Database + "_resume_tokens"
	}
	if r.ResumeTokenCollection == "" {
		r.ResumeTokenCollection = r.Collection + "_resume_tokens"
	}
	r.database = r.client.Database(r.Database)
	r.collection = r.database.Collection(r.Collection)
	r.resumeTokenDatabase = r.client.Database(r.ResumeTokenDatabase)
	r.resumeTokenCollection = r.resumeTokenDatabase.Collection(r.ResumeTokenCollection)
	return r.start(parsedUri)
}

// Close the connection
func (r *ChangeStreamReader) Close() error {
	if r.client != nil {
		r.logger.Info("disconnecting mongodb client")
		return r.client.Disconnect(r.ctx)
	}
	return nil
}

// Read returns the next change event from the change stream.
func (r *ChangeStreamReader) Read() (caddynats.Message, func() error, error) {
	r.logger.Info("waiting for next document in change stream")
	if !r.stream.Next(r.ctx) {
		err := r.stream.Err()
		if strings.Contains(err.Error(), "context canceled") {
			r.logger.Info("stopping mongodb change stream receiver")
			return nil, nil, errors.New("EOF")
		}
		r.logger.Error("error reading from change stream", zap.Error(r.stream.Err()))
		return nil, nil, fmt.Errorf("mongo error: %s", err.Error())
	}
	currentResumeToken, ok := r.stream.Current.Lookup("_id", "_data").StringValueOK()
	if !ok {
		// This should never happen, but we don't want the whole program to crash
		// with a panic if it does.
		return nil, nil, errors.New("could not find resume token in change stream")
	}
	msg, err := r.transform(r.stream.Current)
	if err != nil {
		return nil, nil, err
	}
	return msg, func() error {
		_, err := r.resumeTokenCollection.InsertOne(r.ctx, &resumeToken{Value: currentResumeToken})
		return err
	}, nil
}

// start is used to start the change stream.
func (r *ChangeStreamReader) start(uri *url.URL) error {
	// Find last resume token
	findOneOpts := options.FindOne()
	findOneOpts.SetSort(bson.D{{Key: "_id", Value: -1}})
	lastResumeToken := &resumeToken{}
	err := r.resumeTokenCollection.FindOne(r.ctx, bson.D{}, findOneOpts).Decode(lastResumeToken)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("could not fetch or decode resume token: %v", err)
	}
	// Create change stream options
	changeStreamOpts := options.ChangeStream()
	if lastResumeToken.Value != "" {
		r.logger.Debug("resuming after token", zap.String("token", lastResumeToken.Value))
		changeStreamOpts.SetResumeAfter(bson.D{{Key: "_data", Value: lastResumeToken.Value}})
	}
	// Create change stream
	stream, err := r.collection.Watch(r.ctx, mongo.Pipeline{}, changeStreamOpts)
	if err != nil {
		return err
	}
	r.stream = stream
	r.logger.Info("started mongodb change stream", zap.String("uri", uri.Redacted()))
	return nil
}

// transform is used to transform a change stream document into a message.
func (r *ChangeStreamReader) transform(doc bson.Raw) (caddynats.Message, error) {
	json, err := bson.MarshalExtJSON(r.stream.Current, false, false)
	r.logger.Info("new document received in change stream", zap.String("collection", r.Collection), zap.String("operation", doc.Lookup("operationType").StringValue()))
	if err != nil {
		return nil, err
	}
	msg, err := common.NewChangeStreamEvent(json)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

type resumeToken struct {
	Value string `bson:"value"`
}

var (
	_ caddynats.Reader = (*ChangeStreamReader)(nil)
)
