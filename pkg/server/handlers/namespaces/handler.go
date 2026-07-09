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
	"path"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the namespace handlers
type Handler interface {
	// PostNamespace handles the post namespace request
	PostNamespace(c *gin.Context, req *api.PostNamespaceRequest)
	// ListNamespaces handles the list namespace request
	ListNamespaces(c *gin.Context, req *api.ListNamespaceRequest)
	// GetNamespace handles the get namespace request
	GetNamespace(c *gin.Context, req *api.GetNamespaceRequest)
	// DeleteNamespace handles the delete namespace request
	DeleteNamespace(c *gin.Context, req *api.DeleteNamespaceRequest)
	// PutNamespace handles the put namespace request
	PutNamespace(c *gin.Context, req *api.UpdateNamespaceRequest)
	// HotNamespace handles the hot namespace request
	HotNamespace(c *gin.Context)

	// AddNamespaceMember handles the add namespace member request
	AddNamespaceMember(c *gin.Context, req *api.AddNamespaceMemberRequest)
	// UpdateNamespaceMember handles the update namespace member request
	UpdateNamespaceMember(c *gin.Context, req *api.UpdateNamespaceMemberRequest)
	// DeleteNamespaceMember handles the delete namespace member request
	DeleteNamespaceMember(c *gin.Context, req *api.DeleteNamespaceMemberRequest)
	// ListNamespaceMembers handles the list namespace members request
	ListNamespaceMembers(c *gin.Context, req *api.ListNamespaceMemberRequest)
	// GetNamespaceMemberSelf handles the get self namespace member request
	GetNamespaceMemberSelf(c *gin.Context, req *api.GetNamespaceMemberSelfRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	NsSvc      namespaces.NamespaceService
	Authorizer authz.Authorizer
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		namespaceGroup := e.Group(consts.APIV1 + "/namespaces")

		namespaceGroup.GET("/", server.WrapRequest(h.ListNamespaces))
		namespaceGroup.GET("/:namespace_id", server.WrapRequest(h.GetNamespace))
		namespaceGroup.POST("/", server.WrapRequest(h.PostNamespace))
		namespaceGroup.PUT("/:namespace_id", server.WrapRequest(h.PutNamespace))
		namespaceGroup.DELETE("/:namespace_id", server.WrapRequest(h.DeleteNamespace))
		namespaceGroup.GET("/hot", server.Wrap(h.HotNamespace))

		namespaceGroup.GET("/:namespace_id/members/", server.WrapRequest(h.ListNamespaceMembers))
		namespaceGroup.GET("/:namespace_id/members/self", server.WrapRequest(h.GetNamespaceMemberSelf))
		namespaceGroup.POST("/:namespace_id/members/", server.WrapRequest(h.AddNamespaceMember))
		namespaceGroup.PUT("/:namespace_id/members/:user_id", server.WrapRequest(h.UpdateNamespaceMember))
		namespaceGroup.DELETE("/:namespace_id/members/:user_id", server.WrapRequest(h.DeleteNamespaceMember))

		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
