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

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

func TestDeleteManifest(t *testing.T) {
	logger.SetLevel("debug")
	viper.SetDefault("log.level", "debug")
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	const (
		namespaceName  = "test"
		repositoryName = "test/busybox"
		digestName     = "sha256:2776ee23722eaabcffed77dafd22b7a1da734971bf268a323b6819926dfe1ebd"
		tagName        = "latest"
	)

	ctx := log.Logger.WithContext(context.Background())

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	userObj := &models.User{Provider: enums.ProviderLocal, Username: "head-manifest", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	err = userService.Create(ctx, userObj)
	assert.NoError(t, err)

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	namespaceService := namespaceServiceFactory.New()
	namespaceObj := &models.Namespace{Name: namespaceName, UserID: userObj.ID, Visibility: ptr.Of(enums.VisibilityPrivate)}
	err = namespaceService.Create(ctx, namespaceObj)
	assert.NoError(t, err)

	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	repositoryService := repositoryServiceFactory.New()
	repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName, Visibility: ptr.Of(enums.VisibilityPrivate)}
	err = repositoryService.Create(ctx, repositoryObj)
	assert.NoError(t, err)

	artifactServiceFactory := dao.NewArtifactServiceFactory()
	artifactService := artifactServiceFactory.New()
	artifactObj := &models.Artifact{RepositoryID: repositoryObj.ID, Digest: digestName, Size: 123, ContentType: "application/vnd.oci.image.manifest.v1+json", Raw: []byte(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796","size":800},"layers":[{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904","size":37869}],"annotations":{"category":"Infrastructure","licenses":"Apache-2.0","org.opencontainers.image.authors":"VMware, Inc.","org.opencontainers.image.description":"NGINX Open Source is a web server that can be also used as a reverse proxy, load balancer, and HTTP cache. Recommended for high-demanding sites due to its ability to provide faster content.","org.opencontainers.image.source":"https://github.com/bitnami/charts/tree/main/bitnami/nginx","org.opencontainers.image.title":"nginx","org.opencontainers.image.url":"https://bitnami.com","org.opencontainers.image.version":"15.0.2"}}`)}
	err = artifactService.Create(ctx, artifactObj)
	assert.NoError(t, err)

	tagServiceFactory := dao.NewTagServiceFactory()
	tagService := tagServiceFactory.New()
	tagObj := &models.Tag{RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, Name: tagName}
	err = tagService.Create(ctx, tagObj)
	assert.NoError(t, err)

	h := &handler{
		repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
		tagServiceFactory:        dao.NewTagServiceFactory(),
		artifactServiceFactory:   dao.NewArtifactServiceFactory(),
	}

	// test about delete tag
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	// test about delete artifact by digest
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, digestName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, rec.Code)

	// test about delete artifact by invalid reference
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, "*&invalid-tag"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// test about delete artifact not found the repository
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName+"-none-exist", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// test about delete artifact not found the tag
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName+"-none-exist"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// test about delete artifact not found the digest
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, digestName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.DeleteManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
