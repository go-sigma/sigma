// Copyright 2024 sigma
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

package builders

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/service/builders"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestFactory(t *testing.T) {
	cfg := config.GetConfig()
	enabled := cfg.Daemon.Builder.Enabled
	cfg.Daemon.Builder.Enabled = true
	t.Cleanup(func() {
		cfg.Daemon.Builder.Enabled = enabled
	})

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() builders.Service { return nil }))
	require.NoError(t, digCon.Provide(func() *config.Configuration { return &config.Configuration{} }))
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, factory{}.Initialize(digCon))
	require.NoError(t, digCon.Invoke(func(engine *gin.Engine) {
		routes := engine.Routes()
		require.Len(t, routes, 8)
		require.True(t, hasRoute(routes, http.MethodPost, "/api/v1/namespaces/:namespace_id/repositories/:repository_id/builders/"))
		require.True(t, hasRoute(routes, http.MethodGet, "/api/v1/namespaces/:namespace_id/repositories/:repository_id/builders/:builder_id/runners/:runner_id/log"))
	}))
}

func hasRoute(routes gin.RoutesInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
