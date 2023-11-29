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

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution blob handlers
type Handler interface {
	// DeleteBlob ...
	DeleteBlob(ctx echo.Context) error
	// HeadBlob ...
	HeadBlob(ctx echo.Context) error
	// GetBlob ...
	GetBlob(ctx echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	config                   *configs.Configuration
	authServiceFactory       auth.ServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

type inject struct {
	config                   *configs.Configuration
	authServiceFactory       auth.ServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

// handlerNew creates a new instance of the distribution blob handlers
func handlerNew(injects ...inject) Handler {
	config := configs.GetConfiguration()
	authServiceFactory := auth.NewServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	blobServiceFactory := dao.NewBlobServiceFactory()
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
	}
	return &handler{
		config:                   config,
		authServiceFactory:       authServiceFactory,
		auditServiceFactory:      auditServiceFactory,
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		blobServiceFactory:       blobServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	blobHandler := handlerNew()
	if strings.HasSuffix(urix, "/blobs") {
		switch method {
		case http.MethodGet:
			return blobHandler.GetBlob(c)
		case http.MethodHead:
			return blobHandler.HeadBlob(c)
		default:
			return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 3))
}
