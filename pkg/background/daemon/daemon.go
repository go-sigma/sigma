package daemon

import (
	"fmt"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/infra/registry"
)

// Factory is the interface for the daemon factory
type Factory interface {
	Initialize(c *dig.Container) error
}

// Daemons is the registry for daemon factories
var Daemons = make(registry.Factories[string, Factory])

// Initialize initializes every registered daemon factory, naming the failing factory in the returned error.
func Initialize(digCon *dig.Container) error {
	for name, factory := range Daemons {
		if err := factory.Initialize(digCon); err != nil {
			return fmt.Errorf("failed to initialize daemon factory %q: %v", name, err)
		}
	}

	return nil
}
