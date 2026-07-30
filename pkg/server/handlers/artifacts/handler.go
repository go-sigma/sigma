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

package artifacts

import (
	"path"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/artifacts"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the artifact handlers
type Handler interface {
	// ListArtifact handles the list artifact request
	ListArtifact(c *gin.Context, req *api.ListArtifactRequest)
	// GetArtifact handles the get artifact request
	GetArtifact(c *gin.Context, req *api.GetArtifactRequest)
	// DeleteArtifact handles the delete artifact request
	DeleteArtifact(c *gin.Context, req *api.DeleteArtifactRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	ArtifactSvc artifacts.Service
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		artifactGroup := e.Group(consts.APIV1 + "/namespaces/:namespace_id/artifacts")
		artifactGroup.GET("/", server.WrapRequest(h.ListArtifact))
		artifactGroup.GET("/:digest", server.WrapRequest(h.GetArtifact))
		artifactGroup.DELETE("/:digest", server.WrapRequest(h.DeleteArtifact))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
