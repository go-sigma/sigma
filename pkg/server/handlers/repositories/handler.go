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

package repositories

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
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the repository handlers
type Handler interface {
	// CreateRepository handles the post repository request
	CreateRepository(c *gin.Context, req *api.CreateRepositoryRequest)
	// UpdateRepository handles the put repository request
	UpdateRepository(c *gin.Context, req *api.UpdateRepositoryRequest)
	// GetRepository handles the get repository request
	GetRepository(c *gin.Context, req *api.GetRepositoryRequest)
	// ListRepositories handles the list repository request
	ListRepositories(c *gin.Context, req *api.ListRepositoryRequest)
	// DeleteRepository handles the delete repository request
	DeleteRepository(c *gin.Context, req *api.DeleteRepositoryRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	RepoSvc    svcrepository.RepositoryService
	Authorizer authz.Authorizer
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		repositoryGroup := e.Group(consts.APIV1 + "/namespaces/:namespace_id/repositories")
		repositoryGroup.GET("/", server.WrapRequest(h.ListRepositories))
		repositoryGroup.POST("/", server.WrapRequest(h.CreateRepository))
		repositoryGroup.GET("/:repository_id", server.WrapRequest(h.GetRepository))
		repositoryGroup.PUT("/:repository_id", server.WrapRequest(h.UpdateRepository))
		repositoryGroup.DELETE("/:repository_id", server.WrapRequest(h.DeleteRepository))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
