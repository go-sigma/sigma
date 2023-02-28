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

package blob

import (
	"fmt"
	"regexp"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

// Handlers is the interface for the distribution blob handlers
type Handlers interface {
	DeleteBlob(ctx echo.Context) error
	HeadBlob(ctx echo.Context) error
	GetBlob(ctx echo.Context) error
}

var _ Handlers = &handler{}

var blobRouteReg = regexp.MustCompile(fmt.Sprintf(`^/v2/%s/blobs/%s$`, reference.NameRegexp.String(), digest.DigestRegexp.String()))

type handler struct{}

// New creates a new instance of the distribution blob handlers
func New() Handlers {
	return &handler{}
}
