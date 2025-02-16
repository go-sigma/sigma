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

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the namespace handlers
type Handler interface {
	// PostNamespace handles the post namespace request
	PostNamespace(c echo.Context) error
	// ListNamespaces handles the list namespace request
	ListNamespaces(c echo.Context) error
	// GetNamespace handles the get namespace request
	GetNamespace(c echo.Context) error
	// DeleteNamespace handles the delete namespace request
	DeleteNamespace(c echo.Context) error
	// PutNamespace handles the put namespace request
	PutNamespace(c echo.Context) error
	// HotNamespace handles the hot namespace request
	HotNamespace(c echo.Context) error

	// AddNamespaceMember handles the add namespace member request
	AddNamespaceMember(c echo.Context) error
	// UpdateNamespaceMember handles the update namespace member request
	UpdateNamespaceMember(c echo.Context) error
	// DeleteNamespaceMember handles the delete namespace member request
	DeleteNamespaceMember(c echo.Context) error
	// ListNamespaceMembers handles the list namespace members request
	ListNamespaceMembers(c echo.Context) error
	// GetNamespaceMemberSelf handles the get self namespace member request
	GetNamespaceMemberSelf(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	AuthServiceFactory            auth.AuthServiceFactory
	AuditServiceFactory           dao.AuditServiceFactory
	NamespaceServiceFactory       dao.NamespaceServiceFactory
	NamespaceMemberServiceFactory dao.NamespaceMemberServiceFactory
	RepositoryServiceFactory      dao.RepositoryServiceFactory
	TagServiceFactory             dao.TagServiceFactory
	ArtifactServiceFactory        dao.ArtifactServiceFactory
	ProducerClient                definition.WorkQueueProducer
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	namespaceGroup := e.Group(consts.APIV1 + "/namespaces")

	namespaceHandler := handlerNew(digCon)

	namespaceGroup.GET("/", namespaceHandler.ListNamespaces)
	namespaceGroup.GET("/:namespace_id", namespaceHandler.GetNamespace)
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:namespace_id", namespaceHandler.PutNamespace)
	namespaceGroup.DELETE("/:namespace_id", namespaceHandler.DeleteNamespace)
	namespaceGroup.GET("/hot", namespaceHandler.HotNamespace)

	namespaceGroup.GET("/:namespace_id/members/", namespaceHandler.ListNamespaceMembers)
	namespaceGroup.GET("/:namespace_id/members/self", namespaceHandler.GetNamespaceMemberSelf)
	namespaceGroup.POST("/:namespace_id/members/", namespaceHandler.AddNamespaceMember)
	namespaceGroup.PUT("/:namespace_id/members/:user_id", namespaceHandler.UpdateNamespaceMember)
	namespaceGroup.DELETE("/:namespace_id/members/:user_id", namespaceHandler.DeleteNamespaceMember)

	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
