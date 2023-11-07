package resources

import (
	"context"
)

// resource is an interface that all resources must implement.
// It is not exported because it is only used internally at the moment, but
// it may be exported in the future if it is useful to do so.
type resource interface {
	Configure(ctx context.Context, clients *Clients) error
}
