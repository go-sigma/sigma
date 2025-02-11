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

package caches

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler ...
type Handler interface {
	// CreateCache ...
	CreateCache(c echo.Context) error
	// GetCache ...
	GetCache(c echo.Context) error
	// DeleteCache ...
	DeleteCache(c echo.Context) error
}

type handler struct {
	digCon *dig.Container
}

// handlerNew creates a new instance of the builder handlers
func handlerNew(c *dig.Container) Handler {
	return &handler{
		digCon: c,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	handler := handlerNew(digCon)
	echo := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	cacheGroup := echo.Group(consts.APIV1 + "/caches")
	cacheGroup.POST("/:builder_id", handler.CreateCache)
	cacheGroup.GET("/:builder_id", handler.GetCache)
	cacheGroup.DELETE("/:builder_id", handler.DeleteCache)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
