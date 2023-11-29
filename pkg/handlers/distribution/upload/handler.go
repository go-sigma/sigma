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

package upload

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution blob handlers
type Handler interface {
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

var _ Handler = &handler{}

type handler struct {
	config                   *configs.Configuration
	authServiceFactory       auth.ServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	blobUploadServiceFactory dao.BlobUploadServiceFactory
}

type inject struct {
	config                   *configs.Configuration
	authServiceFactory       auth.ServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	blobUploadServiceFactory dao.BlobUploadServiceFactory
}

// handlerNew creates a new instance of the distribution upload blob handlers
func handlerNew(injects ...inject) Handler {
	config := configs.GetConfiguration()
	authServiceFactory := auth.NewServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	blobServiceFactory := dao.NewBlobServiceFactory()
	blobUploadServiceFactory := dao.NewBlobUploadServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.config != nil {
			config = ij.config
		}
		if ij.authServiceFactory != nil {
			authServiceFactory = ij.authServiceFactory
		}
		if ij.auditServiceFactory != nil {
			auditServiceFactory = ij.auditServiceFactory
		}
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.repositoryServiceFactory != nil {
			repositoryServiceFactory = ij.repositoryServiceFactory
		}
		if ij.blobServiceFactory != nil {
			blobServiceFactory = ij.blobServiceFactory
		}
		if ij.blobUploadServiceFactory != nil {
			blobUploadServiceFactory = ij.blobUploadServiceFactory
		}
	}
	return &handler{
		config:                   config,
		authServiceFactory:       authServiceFactory,
		auditServiceFactory:      auditServiceFactory,
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		blobServiceFactory:       blobServiceFactory,
		blobUploadServiceFactory: blobUploadServiceFactory,
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
