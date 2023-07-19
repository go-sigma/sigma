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

	"github.com/go-sigma/sigma/pkg/consts"
	rhandlers "github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the system handlers
type Handlers interface {
	// GetEndpoint handles the get endpoint request
	GetEndpoint(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

type inject struct{}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	return &handlers{}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	validatorGroup := e.Group(consts.APIV1 + "/systems")
	repositoryHandler := handlerNew()
	validatorGroup.GET("/endpoint", repositoryHandler.GetEndpoint)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}