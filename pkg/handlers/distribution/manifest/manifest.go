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

package manifest

import "github.com/labstack/echo/v4"

// Handlers is the interface for the distribution manifest handlers
type Handlers interface {
	GetManifest(ctx echo.Context) error
	HeadManifest(ctx echo.Context) error
	PutManifest(ctx echo.Context) error
	DeleteManifest(ctx echo.Context) error
}

var _ Handlers = &handler{}

type handler struct{}

// New creates a new instance of the distribution manifest handlers
func New() Handlers {
	return &handler{}
}
