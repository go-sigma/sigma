package registry

import (
	"fmt"
)

type Factories[K ~string, T any] map[K]T

// Register registers a new factory
func (f Factories[K, T]) Register(name K, factory T) error {
	if _, ok := f[name]; ok {
		return fmt.Errorf("factory %q already registered", name)
	}
	f[name] = factory
	return nil
}
