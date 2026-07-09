package daemon

import (
	"fmt"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/infra/registry"
	"github.com/go-sigma/sigma/pkg/validators"
)

// Factory is the interface for the daemon factory
type Factory interface {
	Initialize(c *dig.Container) error
}

// Daemons is the registry for daemon factories
var Daemons = make(registry.Factories[string, Factory])

// Initialize ...
func Initialize(digCon *dig.Container) error {
	err := validators.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize validators: %v", err)
	}

	for name, factory := range Daemons {
		if err := factory.Initialize(digCon); err != nil {
			return fmt.Errorf("failed to initialize daemon factory %q: %v", name, err)
		}
	}

	return nil
}
