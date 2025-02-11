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

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the system handlers
type Handler interface {
	// GetEndpoint handles the get endpoint request
	GetEndpoint(c echo.Context) error
	// GetVersion handles the get version request
	GetVersion(c echo.Context) error
	// GetConfig handles the get config request
	GetConfig(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	config configs.Configuration
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return &handler{
		config: utils.MustGetObjFromDigCon[configs.Configuration](digCon),
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	group := e.Group(consts.APIV1 + "/systems")
	handler := handlerNew(digCon)
	group.GET("/endpoint", handler.GetEndpoint)
	group.GET("/version", handler.GetVersion)
	group.GET("/config", handler.GetConfig)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
