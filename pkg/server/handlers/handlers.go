// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/infra/registry"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/validators"
)

// Factory is the interface for the storage router factory
type Factory interface {
	Initialize(c *dig.Container) error
}

// Routers is the registry for storage router factories
var Routers = make(registry.Factories[string, Factory])

// Initialize ...
func Initialize(digCon *dig.Container) error {
	err := validators.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize validators: %v", err)
	}

	for name, factory := range Routers {
		if err := factory.Initialize(digCon); err != nil {
			return fmt.Errorf("failed to initialize router factory %q: %v", name, err)
		}
	}

	return nil
}

// InitializeDistribution ...
func InitializeDistribution(digCon *dig.Container) error {
	if err := digCon.Invoke(func(e *gin.Engine) error {
		e.Any("/v2/*path", server.Wrap(func(c *gin.Context) {
			if err := distribution.All(c, digCon); err != nil {
				slog.Error("handle distribution request failed", "err", err, "method", c.Request.Method, "path", c.Request.URL.Path)
			}
		}))
		return nil
	}); err != nil {
		return fmt.Errorf("failed to initialize distribution handlers: %v", err)
	}
	return nil
}
