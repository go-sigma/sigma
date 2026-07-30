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

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/coderepos"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the system handlers
type Handler interface {
	// List list all of the code repositories
	List(c *gin.Context, req *api.ListCodeRepositoryRequest)
	// Get get code repository by id
	Get(c *gin.Context, req *api.GetCodeRepositoryRequest)
	// ListOwner list all of the code repository owner
	ListOwners(c *gin.Context, req *api.ListCodeRepositoryOwnerRequest)
	// ListBranches ...
	ListBranches(c *gin.Context, req *api.ListCodeRepositoryBranchesRequest)
	// GetBranch ...
	GetBranch(c *gin.Context, req *api.GetCodeRepositoryBranchRequest)
	// Resync resync all of the code repositories
	Resync(c *gin.Context, req *api.GetCodeRepositoryResyncRequest)
	// Providers get providers
	Providers(c *gin.Context)
	// User3rdParty get user 3rdparty
	User3rdParty(c *gin.Context, req *api.GetCodeRepositoryUser3rdPartyRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	CodeRepoSvc coderepos.Service
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		config := config.GetConfig()
		if config.Daemon.Builder.Enabled { // TODO: use dig
			codereposGroup := e.Group(consts.APIV1 + "/coderepos")
			codereposGroup.GET("/providers", server.Wrap(h.Providers))
			codereposGroup.GET("/:provider", server.WrapRequest(h.List))
			codereposGroup.GET("/:provider/repos/:id", server.WrapRequest(h.Get))
			codereposGroup.GET("/:provider/user3rdparty", server.WrapRequest(h.User3rdParty))
			codereposGroup.GET("/:provider/resync", server.WrapRequest(h.Resync))
			codereposGroup.GET("/:provider/owners", server.WrapRequest(h.ListOwners))
			codereposGroup.GET("/:provider/repos/:id/branches", server.WrapRequest(h.ListBranches))
			codereposGroup.GET("/:provider/repos/:id/branches/:name", server.WrapRequest(h.GetBranch))
		}
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
