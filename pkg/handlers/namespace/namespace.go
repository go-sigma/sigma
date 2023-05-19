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

package namespace

import "github.com/labstack/echo/v4"

// Handlers is the interface for the namespace handlers
type Handlers interface {
	// PostNamespace handles the post namespace request
	PostNamespace(c echo.Context) error
	// ListNamespace handles the list namespace request
	ListNamespace(c echo.Context) error
	// DeleteNamespace handles the delete namespace request
	DeleteNamespace(c echo.Context) error
	// PutNamespace handles the put namespace request
	PutNamespace(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct{}

// New creates a new instance of the distribution handlers
func New() Handlers {
	return &handlers{}
}
