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

package upload

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handlers is the interface for the distribution blob handlers
type Handlers interface {
	// DeleteUpload ...
	DeleteUpload(ctx echo.Context) error
	// GetUpload ...
	GetUpload(ctx echo.Context) error
	// PatchUpload ...
	PatchUpload(ctx echo.Context) error
	// PostUpload ...
	PostUpload(ctx echo.Context) error
	// PutUpload ...
	PutUpload(ctx echo.Context) error
}

var _ Handlers = &handler{}

type handler struct {
	blobUploadServiceFactory dao.BlobUploadServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

type inject struct {
	blobUploadServiceFactory dao.BlobUploadServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

// handlerNew creates a new instance of the distribution upload blob handlers
func handlerNew(injects ...inject) Handlers {
	blobUploadServiceFactory := dao.NewBlobUploadServiceFactory()
	blobServiceFactory := dao.NewBlobServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.blobUploadServiceFactory != nil {
			blobUploadServiceFactory = ij.blobUploadServiceFactory
		}
		if ij.blobServiceFactory != nil {
			blobServiceFactory = ij.blobServiceFactory
		}
	}
	return &handler{
		blobUploadServiceFactory: blobUploadServiceFactory,
		blobServiceFactory:       blobServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context) error {
	method := c.Request().Method
	uri := c.Request().RequestURI

	blobUploadHandler := handlerNew()
	if method == http.MethodPost && strings.HasSuffix(uri, "blobs/uploads/") {
		return blobUploadHandler.PostUpload(c)
	}

	urix := uri[:strings.LastIndex(uri, "/")]
	if strings.HasSuffix(urix, "/blobs/uploads") {
		switch method {
		case http.MethodGet:
			return blobUploadHandler.GetUpload(c)
		case http.MethodPatch:
			return blobUploadHandler.PatchUpload(c)
		case http.MethodPut:
			return blobUploadHandler.PutUpload(c)
		case http.MethodDelete:
			return blobUploadHandler.DeleteUpload(c)
		default:
			return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 2))
}
