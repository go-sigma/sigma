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

package coderepos

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the system handlers
type Handler interface {
	// List list all of the code repositories
	List(c echo.Context) error
	// Get get code repository by id
	Get(c echo.Context) error
	// ListOwner list all of the code repository owner
	ListOwners(c echo.Context) error
	// ListBranches ...
	ListBranches(c echo.Context) error
	// GetBranch ...
	GetBranch(c echo.Context) error
	// Resync resync all of the code repositories
	Resync(c echo.Context) error
	// Providers get providers
	Providers(c echo.Context) error
	// User3rdParty get user 3rdparty
	User3rdParty(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	NamespaceServiceFactory      dao.NamespaceServiceFactory
	RepositoryServiceFactory     dao.RepositoryServiceFactory
	CodeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
	UserServiceFactory           dao.UserServiceFactory
	AuditServiceFactory          dao.AuditServiceFactory
	BuilderServiceFactory        dao.BuilderServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	codeRepositoryHandler := handlerNew(digCon)

	config := configs.GetConfiguration()
	if config.Daemon.Builder.Enabled { // TODO: use dig
		codereposGroup := e.Group(consts.APIV1 + "/coderepos")
		codereposGroup.GET("/providers", codeRepositoryHandler.Providers)
		codereposGroup.GET("/:provider", codeRepositoryHandler.List)
		codereposGroup.GET("/:provider/repos/:id", codeRepositoryHandler.Get)
		codereposGroup.GET("/:provider/user3rdparty", codeRepositoryHandler.User3rdParty)
		codereposGroup.GET("/:provider/resync", codeRepositoryHandler.Resync)
		codereposGroup.GET("/:provider/owners", codeRepositoryHandler.ListOwners)
		codereposGroup.GET("/:provider/repos/:id/branches", codeRepositoryHandler.ListBranches)
		codereposGroup.GET("/:provider/repos/:id/branches/:name", codeRepositoryHandler.GetBranch)
	}
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
