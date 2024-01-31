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

package manifest

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

// Handler is the interface for the distribution manifest handlers
type Handler interface {
	// GetManifest ...
	GetManifest(ctx echo.Context) error
	// HeadManifest ...
	HeadManifest(ctx echo.Context) error
	// PutManifest ...
	PutManifest(ctx echo.Context) error
	// DeleteManifest ...
	DeleteManifest(ctx echo.Context) error
	// GetReferrer ...
	GetReferrer(ctx echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	config                   *configs.Configuration
	authServiceFactory       auth.AuthServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

type inject struct {
	config                   *configs.Configuration
	authServiceFactory       auth.AuthServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
}

// New creates a new instance of the distribution manifest handlers
func handlerNew(injects ...inject) Handler {
	config := configs.GetConfiguration()
	authServiceFactory := auth.NewAuthServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	tagServiceFactory := dao.NewTagServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
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
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
		if ij.tagServiceFactory != nil {
			tagServiceFactory = ij.tagServiceFactory
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
		artifactServiceFactory:   artifactServiceFactory,
		tagServiceFactory:        tagServiceFactory,
		blobServiceFactory:       blobServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	manifestHandler := handlerNew()
	if strings.HasSuffix(urix, "/manifests") {
		switch method {
		case http.MethodGet:
			return manifestHandler.GetManifest(c)
		case http.MethodHead:
			return manifestHandler.HeadManifest(c)
		case http.MethodPut:
			return manifestHandler.PutManifest(c)
		case http.MethodDelete:
			return manifestHandler.DeleteManifest(c)
		default:
			return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	} else if strings.HasSuffix(urix, "/referrers") && method == http.MethodGet {
		return manifestHandler.GetReferrer(c)
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 4))
}
