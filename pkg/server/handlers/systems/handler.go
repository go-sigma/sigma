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

package systems

import (
	"path"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/systems"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the system handlers
type Handler interface {
	// GetEndpoint handles the get endpoint request
	GetEndpoint(c *gin.Context)
	// GetVersion handles the get version request
	GetVersion(c *gin.Context)
	// GetConfig handles the get config request
	GetConfig(c *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	SystemsSvc systems.SystemsService
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		group := e.Group(consts.APIV1 + "/systems")
		group.GET("/endpoint", server.Wrap(h.GetEndpoint))
		group.GET("/version", server.Wrap(h.GetVersion))
		group.GET("/config", server.Wrap(h.GetConfig))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
