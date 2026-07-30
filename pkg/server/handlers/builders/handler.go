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

package builders

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
	"github.com/go-sigma/sigma/pkg/service/builders"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the builder handlers
type Handler interface {
	// CreateBuilder handles the create builder request
	CreateBuilder(c *gin.Context, req *api.CreateBuilderRequest)
	// UpdateBuilder handles the update builder request
	UpdateBuilder(c *gin.Context, req *api.UpdateBuilderRequest)
	// ListRunners handles the list builder runners request
	ListRunners(c *gin.Context, req *api.ListBuilderRunnersRequest)
	// PostRunnerRun ...
	PostRunnerRun(c *gin.Context, req *api.PostRunnerRun)
	// GetRunnerRerun ...
	GetRunnerRerun(c *gin.Context, req *api.GetRunnerStop)
	// GetRunnerStop ...
	GetRunnerStop(c *gin.Context, req *api.GetRunnerStop)
	// GetRunnerLog ...
	GetRunnerLog(c *gin.Context, req *api.GetRunnerLog)
	// GetRunner ...
	GetRunner(c *gin.Context, req *api.GetRunner)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	BuilderSvc builders.Service
	Config     *config.Configuration
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		config := config.GetConfig() // TODO: use dig
		if config.Daemon.Builder.Enabled {
			builderGroup := e.Group(consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/builders")
			builderGroup.POST("/", server.WrapRequest(h.CreateBuilder))
			builderGroup.PUT("/:builder_id", server.WrapRequest(h.UpdateBuilder))
			builderGroup.GET("/:builder_id/runners/", server.WrapRequest(h.ListRunners))
			builderGroup.POST("/:builder_id/runners/run", server.WrapRequest(h.PostRunnerRun))
			builderGroup.GET("/:builder_id/runners/:runner_id", server.WrapRequest(h.GetRunner))
			builderGroup.GET("/:builder_id/runners/:runner_id/stop", server.WrapRequest(h.GetRunnerStop))
			builderGroup.GET("/:builder_id/runners/:runner_id/rerun", server.WrapRequest(h.GetRunnerRerun))
			builderGroup.GET("/:builder_id/runners/:runner_id/log", server.WrapRequest(h.GetRunnerLog))
		}
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
