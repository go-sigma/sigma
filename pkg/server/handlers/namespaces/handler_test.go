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

package namespaces

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server"
	svcnamespace "github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestInitialize(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() svcnamespace.Service { return nil }))
	require.NoError(t, digCon.Provide(func() authz.Authorizer { return nil }))
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, digCon.Invoke(Initialize))
	require.NoError(t, digCon.Invoke(func(engine *gin.Engine) {
		routes := engine.Routes()
		require.Len(t, routes, 11)
		require.True(t, hasRoute(routes, http.MethodPost, "/api/v1/namespaces/"))
		require.True(t, hasRoute(routes, http.MethodGet, "/api/v1/namespaces/:namespace_id/members/self"))
		require.True(t, hasRoute(routes, http.MethodDelete, "/api/v1/namespaces/:namespace_id/members/:user_id"))
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

func newNamespacesTestRouter(h handler, user *models.User) *gin.Engine {
	router := testkit.NewGin()
	router.Use(func(c *gin.Context) {
		if user != nil {
			c.Set(consts.ContextUser, user)
		}
	})
	group := router.Group(consts.APIV1 + "/namespaces")
	group.GET("/", server.WrapRequest(h.ListNamespaces))
	group.GET("/:namespace_id", server.WrapRequest(h.GetNamespace))
	group.POST("/", server.WrapRequest(h.PostNamespace))
	group.PUT("/:namespace_id", server.WrapRequest(h.PutNamespace))
	group.DELETE("/:namespace_id", server.WrapRequest(h.DeleteNamespace))
	group.GET("/hot", server.Wrap(h.HotNamespace))
	return router
}

type fakeAuthorizer struct {
	namespace func(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error)
}

var _ authz.Authorizer = fakeAuthorizer{}

func (f fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Namespace(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error) {
	if f.namespace == nil {
		return false, nil
	}
	return f.namespace(ctx, user, namespaceID, auth)
}

func (f fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) Repository(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
