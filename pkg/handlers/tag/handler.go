// Copyright 2023 XImager
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

package tag

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	rhandlers "github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/utils"
)

// Handlers is the interface for the tag handlers
type Handlers interface {
	// ListTag handles the list tag request
	ListTag(c echo.Context) error
	// GetTag handles the get tag request
	GetTag(c echo.Context) error
	// DeleteTag handles the delete tag request
	DeleteTag(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// handlerNew creates a new instance of the distribution handlers
func handlerNew() (Handlers, error) {
	return &handlers{}, nil
}

type factory struct{}

func (f factory) Initialize(e *echo.Echo) error {
	tagGroup := e.Group("/namespace/:namespace/tag")
	tagHandler, err := handlerNew()
	if err != nil {
		return err
	}
	tagGroup.GET("/", tagHandler.ListTag)
	tagGroup.GET("/:id", tagHandler.GetTag)
	tagGroup.DELETE("/:id", tagHandler.DeleteTag)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(handlers{}).PkgPath()), &factory{}))
}
