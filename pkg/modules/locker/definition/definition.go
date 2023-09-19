package definition

import (
	"context"
	"time"
)

// Lock ...
type Lock interface {
	// Unlock ...
	Unlock() error
}

// Locker ...
type Locker interface {
	// Lock ...
	Lock(ctx context.Context, name string, expire time.Duration) (Lock, error)
}
