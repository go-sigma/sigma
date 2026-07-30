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

package blob

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcblob "github.com/go-sigma/sigma/pkg/service/distribution/blob"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution blob handlers
type Handler interface {
	// DeleteBlob ...
	DeleteBlob(ctx *gin.Context)
	// HeadBlob ...
	HeadBlob(ctx *gin.Context)
	// GetBlob ...
	GetBlob(ctx *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config     *config.Configuration
	BlobSvc    svcblob.Service
	Authorizer authz.Authorizer
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c *gin.Context, digCon *dig.Container) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	return digCon.Invoke(func(h handler) error {
		if strings.HasSuffix(urix, "/blobs") {
			switch method {
			case http.MethodGet:
				h.GetBlob(c)
				return nil
			case http.MethodHead:
				h.HeadBlob(c)
				return nil
			default:
				c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
				return nil
			}
		}
		return distribution.ErrNext
	})
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 3))
}

// dsError maps a service error (typically an errcode.ErrCode) into a
// distribution-spec HTTP error response.
func (h *handler) dsError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
		errcode.NewDSError(c, e)
		return
	}
	errcode.NewDSError(c, errcode.DSErrCodeUnknown)
}
