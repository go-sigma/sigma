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

	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers/analytics"
	"github.com/go-sigma/sigma/pkg/server/handlers/artifacts"
	"github.com/go-sigma/sigma/pkg/server/handlers/builders"
	"github.com/go-sigma/sigma/pkg/server/handlers/coderepos"
	"github.com/go-sigma/sigma/pkg/server/handlers/daemons"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/server/handlers/namespaces"
	"github.com/go-sigma/sigma/pkg/server/handlers/oauth2"
	"github.com/go-sigma/sigma/pkg/server/handlers/repositories"
	"github.com/go-sigma/sigma/pkg/server/handlers/systems"
	"github.com/go-sigma/sigma/pkg/server/handlers/tags"
	handlertokens "github.com/go-sigma/sigma/pkg/server/handlers/tokens"
	"github.com/go-sigma/sigma/pkg/server/handlers/users"
	handlervalidators "github.com/go-sigma/sigma/pkg/server/handlers/validators"
	"github.com/go-sigma/sigma/pkg/server/handlers/webhooks"
	"github.com/go-sigma/sigma/pkg/validators"
)

// Initialize registers all API handler routes.
func Initialize(digCon *dig.Container) error {
	err := validators.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize validators: %v", err)
	}

	for _, item := range []struct {
		name       string
		initialize any
	}{
		{name: "analytics", initialize: analytics.Initialize},
		{name: "artifacts", initialize: artifacts.Initialize},
		{name: "builders", initialize: builders.Initialize},
		{name: "coderepos", initialize: coderepos.Initialize},
		{name: "daemons", initialize: daemons.Initialize},
		{name: "namespaces", initialize: namespaces.Initialize},
		{name: "oauth2", initialize: oauth2.Initialize},
		{name: "repositories", initialize: repositories.Initialize},
		{name: "systems", initialize: systems.Initialize},
		{name: "tags", initialize: tags.Initialize},
		{name: "tokens", initialize: handlertokens.Initialize},
		{name: "users", initialize: users.Initialize},
		{name: "validators", initialize: handlervalidators.Initialize},
		{name: "webhooks", initialize: webhooks.Initialize},
	} {
		if err := digCon.Invoke(item.initialize); err != nil {
			return fmt.Errorf("failed to initialize handler %q: %v", item.name, err)
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
