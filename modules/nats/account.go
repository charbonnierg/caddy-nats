package nats

// import (
// 	"context"
// 	"encoding/json"

// 	"github.com/nats-io/jwt/v2"
// 	"github.com/nats-io/nats.go"
// )

// type Resources struct {
// 	StreamsRaw      []json.RawMessage `json:"streams,omitempty"`
// 	BucketsRaw      []json.RawMessage `json:"buckets,omitempty"`
// 	ObjectStoresRaw []json.RawMessage `json:"object_stores,omitempty"`
// 	BridgesRaw      []json.RawMessage `json:"bridges,omitempty"`
// 	AuthPoliciesRaw []json.RawMessage `json:"auth_policies,omitempty"`
// }

// // Account is an an interface that allows retrieving
// // the options used to connect to a NATS server.
// type Account struct {
// 	Resources *Resources

// 	nc *nats.Conn
// 	// Not required when using beyond deployment
// 	// ConnectOptions []nats.Option
// 	// Provision resources
// 	bridges      []Bridge
// 	streams      []Stream
// 	buckets      []Bucket
// 	objectStores []ObjectStore
// 	authPolicies []AuthPolicy
// }

// func (a *Account) Provision(app App) error {
// 	a.Bridges = []Bridge{}
// 	a.Streams = []Stream{}
// 	a.Buckets = []Bucket{}
// 	a.ObjectStores = []ObjectStore{}
// 	a.AuthPolicies = []AuthPolicy{}
// 	return nil
// }

// func (a *Account) CreateStreams() error {
// 	for _, stream := range a.Streams {
// 		cfg, err := stream.StreamConfig()
// 		if err != nil {
// 			return err
// 		}
// 		opts := stream.JetStreamOptions()
// 		js, err := a.nc.JetStream(opts...)
// 		if err != nil {
// 			return err
// 		}
// 		if _, err := js.AddStream(cfg); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (a *Account) CreateBuckets() error {
// 	return nil
// }

// func (a *Account) CreateObjectStores() error {
// 	return nil
// }

// func (a *Account) CreateBridges() error {
// 	return nil
// }

// func (a *Account) CreateAll() error {
// 	if err := a.CreateStreams(); err != nil {
// 		return err
// 	}
// 	if err := a.CreateBuckets(); err != nil {
// 		return err
// 	}
// 	if err := a.CreateObjectStores(); err != nil {
// 		return err
// 	}
// 	if err := a.CreateBridges(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Stream is an interface that allows retrieving
// // the options used to create a NATS stream.
// type Stream interface {
// 	StreamConfig() (*nats.StreamConfig, error)
// 	JetStreamOptions() []nats.JSOpt
// }

// // Bucket is an interface that allows retrieving
// // the options used to create a NATS bucket.
// type Bucket interface {
// 	BucketConfig() (*nats.KeyValueConfig, error)
// 	JetStreamOptions() []nats.JSOpt
// }

// // ObjectStore is an interface that allows retrieving
// // the options used to create a NATS object store.
// type ObjectStore interface {
// 	StoreConfig() (*nats.ObjectStoreConfig, error)
// 	JetStreamOptions() []nats.JSOpt
// }

// // Bridge is an interface that let data flow from a source to a destination.
// type Bridge interface {
// 	Start() error
// 	Stop() error
// }

// type Request interface {
// 	Claims() *jwt.AuthorizationRequestClaims
// 	Context() context.Context
// }

// type Matcher interface {
// 	Match(request Request) bool
// }

// type Issuer interface {
// 	Issue(claims Request) (*jwt.UserClaims, error)
// }

// type AuthPolicy interface {
// 	Matchers() []Matcher
// 	Issuer() Issuer
// }
