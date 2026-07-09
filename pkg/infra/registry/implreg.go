// Package implreg provides a generic key→factory registry that replaces the
// ad-hoc map[string]Factory pattern used throughout the codebase.
//
// Typical usage:
//
//	// declaration (replaces: var DriverFactories = make(map[string]Factory))
//	var DriverFactories = registry.New[Factory]()
//
//	// registration in init() (replaces: DriverFactories[key] = &factory{})
//	func init() {
//	    registry.MustRegister(DriverFactories, "mydriver", &factory{})
//	}
//
//	// lookup (replaces: factory, ok := DriverFactories[key])
//	factory, err := DriverFactories.Get("mydriver")
package registry

import (
	"fmt"
	"sync"
)

// Registry is a generic key → implementation registry backed by sync.Map.
// F is typically a factory interface.
type Registry[F any] struct {
	m sync.Map
}

// New creates an empty Registry.
func New[F any]() *Registry[F] {
	return &Registry[F]{}
}

// Register adds f under key. Returns an error if key is already registered.
func (r *Registry[F]) Register(key string, f F) error {
	if _, loaded := r.m.LoadOrStore(key, f); loaded {
		return fmt.Errorf("registry: %q already registered", key)
	}
	return nil
}

// Get returns the implementation registered under key, or an error if not found.
func (r *Registry[F]) Get(key string) (F, error) {
	v, ok := r.m.Load(key)
	if !ok {
		var zero F
		return zero, fmt.Errorf("registry: %q not registered", key)
	}
	return v.(F), nil
}

// MustRegister adds f under key and panics if the key is already registered.
// Intended for use in init() where duplicate registration is a programming error.
func (r *Registry[F]) MustRegister(key string, f F) {
	if err := r.Register(key, f); err != nil {
		panic(err)
	}
}

// MustRegister is a package-level convenience that calls r.MustRegister.
func MustRegister[F any](r *Registry[F], key string, f F) {
	r.MustRegister(key, f)
}
