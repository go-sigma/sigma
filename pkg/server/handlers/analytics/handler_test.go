// Copyright 2026 sigma
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

package analytics

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestInitialize(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() svcanalytics.Service { return nil }))
	require.NoError(t, digCon.Provide(func() authz.Authorizer { return nil }))
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, digCon.Invoke(Initialize))
	require.NoError(t, digCon.Invoke(func(engine *gin.Engine) {
		requireRoutes(t, engine, map[string]string{
			"/api/v1/users/:user_id/activity/heatmap":          http.MethodGet,
			"/api/v1/namespaces/:namespace_id/activity/trends": http.MethodGet,
		})
	}))
}

func requireRoutes(t *testing.T, engine *gin.Engine, expected map[string]string) {
	t.Helper()
	routes := make(map[string]string, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[route.Path] = route.Method
	}
	for path, method := range expected {
		require.Equal(t, method, routes[path], path)
	}
}
