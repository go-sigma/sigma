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

package tags

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
	"github.com/go-sigma/sigma/pkg/service/tags"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// ListTag handles the list tag request
	ListTag(c *gin.Context, req *api.ListTagRequest)
	// GetTag handles the get tag request
	GetTag(c *gin.Context, req *api.GetTagRequest)
	// DeleteTag handles the delete tag request
	DeleteTag(c *gin.Context, req *api.DeleteTagRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	TagSvc     tags.TagService
	Authorizer authz.Authorizer
}

type factory struct{}

func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		tagGroup := e.Group(consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/tags")
		tagGroup.GET("/", server.WrapRequest(h.ListTag))
		tagGroup.GET("/:id", server.WrapRequest(h.GetTag))
		tagGroup.DELETE("/:id", server.WrapRequest(h.DeleteTag))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
