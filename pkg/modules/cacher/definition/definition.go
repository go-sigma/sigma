package definition

import (
	"context"
	"fmt"
)

var (
	ErrValNil   = fmt.Errorf("Val should not be nil")
	ErrNotFound = fmt.Errorf("Key not found")
)

// Fetcher ...
type Fetcher[T any] func(key string) (T, error)

// Cacher ...
type Cacher[T any] interface {
	// Set sets the value of given key if it is new to the cache.
	// Param val should not be nil.
	Set(ctx context.Context, key string, val T) error
	// Get tries to fetch a value corresponding to the given key from the cache.
	// If error occurs during the first time fetching, it will be cached until the
	// sequential fetching triggered by the refresh goroutine succeed.
	Get(ctx context.Context, key string) (T, error)
	// Del deletes the value corresponding to the given key from the cache.
	Del(ctx context.Context, key string) error
}

// CacherFactory ...
type CacherFactory[T any] interface {
	New(prefix string, fetcher Fetcher[T]) (Cacher[T], error)
}
