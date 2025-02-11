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

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/go-sigma/sigma/pkg/auth"
// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// )

// func TestDeleteManifest(t *testing.T) {
// 	logger.SetLevel("debug")
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 		digestName     = "sha256:2776ee23722eaabcffed77dafd22b7a1da734971bf268a323b6819926dfe1ebd" // nolint: gosec
// 		tagName        = "latest"
// 	)

// 	ctx := log.Logger.WithContext(context.Background())

// 	userObj := &models.User{Username: "head-manifest", Password: ptr.Of("test"), Role: enums.UserRoleRoot, Email: ptr.Of("test@gmail.com")}
// 	assert.NoError(t, dao.NewUserServiceFactory().New().Create(ctx, userObj))

// 	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, dao.NewNamespaceServiceFactory().New().Create(ctx, namespaceObj))

// 	repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName}
// 	assert.NoError(t, dao.NewRepositoryServiceFactory().New().Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

// 	artifactObj := &models.Artifact{NamespaceID: namespaceObj.ID, RepositoryID: repositoryObj.ID, Digest: digestName, Size: 123, ContentType: "application/vnd.oci.image.manifest.v1+json", Raw: []byte(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796","size":800},"layers":[{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904","size":37869}],"annotations":{"category":"Infrastructure","licenses":"Apache-2.0","org.opencontainers.image.authors":"VMware, Inc.","org.opencontainers.image.description":"NGINX Open Source is a web server that can be also used as a reverse proxy, load balancer, and HTTP cache. Recommended for high-demanding sites due to its ability to provide faster content.","org.opencontainers.image.source":"https://github.com/bitnami/charts/tree/main/bitnami/nginx","org.opencontainers.image.title":"nginx","org.opencontainers.image.url":"https://bitnami.com","org.opencontainers.image.version":"15.0.2"}}`)}
// 	assert.NoError(t, dao.NewArtifactServiceFactory().New().Create(ctx, artifactObj))

// 	tagObj := &models.Tag{RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, Name: tagName}
// 	assert.NoError(t, dao.NewTagServiceFactory().New().Create(ctx, tagObj))

// 	h := &handler{
// 		authServiceFactory:       auth.NewAuthServiceFactory(),
// 		namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
// 		repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
// 		tagServiceFactory:        dao.NewTagServiceFactory(),
// 		artifactServiceFactory:   dao.NewArtifactServiceFactory(),
// 	}

// 	// test about delete tag
// 	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err := h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusAccepted, rec.Code)

// 	// test about delete artifact by digest
// 	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, digestName), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusAccepted, rec.Code)

// 	// test about delete artifact by invalid reference
// 	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, "*&invalid-tag"), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)

// 	// test about delete artifact not found the repository
// 	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName+"-none-exist", tagName), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNotFound, rec.Code)

// 	// test about delete artifact not found the tag
// 	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName+"-none-exist"), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNotFound, rec.Code)

// 	// test about delete artifact not found the digest
// 	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, digestName), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = h.DeleteManifest(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNotFound, rec.Code)
// }
