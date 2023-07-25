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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestHeadManifestFallbackProxy(t *testing.T) {
	logger.SetLevel("debug")

	mux := http.NewServeMux()

	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/library/busybox/manifest/sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.index.v1+json")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/library/busybox/manifest/sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.index.v1+json")
		w.WriteHeader(http.StatusInternalServerError)
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	viper.Reset()
	viper.SetDefault("log.level", "info")
	viper.SetDefault("proxy.endpoint", s.URL)
	viper.SetDefault("proxy.tlsVerify", true)

	handler := &handler{}

	req := httptest.NewRequest(http.MethodHead, "/v2/library/busybox/manifest/sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err := handler.headManifestFallbackProxy(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodHead, "/v2/library/busybox/manifest/sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = handler.headManifestFallbackProxy(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHeadManifestFallbackProxyAuthError(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	h := &handler{}

	// test about proxy server auth internal server error
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/library/busybox/manifests/%s", "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err := h.headManifestFallbackProxy(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHeadManifest(t *testing.T) {
	viper.SetDefault("log.level", "debug")
	logger.SetLevel("debug")
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

	ctx := log.Logger.WithContext(context.Background())

	const (
		namespaceName  = "test"
		repositoryName = "test/busybox"
		digestName     = "sha256:2776ee23722eaabcffed77dafd22b7a1da734971bf268a323b6819926dfe1ebd"
		tagName        = "latest"
	)

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	userObj := &models.User{Provider: enums.ProviderLocal, Username: "head-manifest", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	err = userService.Create(ctx, userObj)
	assert.NoError(t, err)

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	namespaceService := namespaceServiceFactory.New()
	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
	err = namespaceService.Create(ctx, namespaceObj)
	assert.NoError(t, err)

	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	repositoryService := repositoryServiceFactory.New()
	repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName, Visibility: enums.VisibilityPrivate}
	err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID})
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

	mux := http.NewServeMux()

	s := httptest.NewServer(mux)
	defer s.Close()

	handler := handlerNew()
	assert.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, digestName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err = handler.HeadManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = handler.HeadManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// test about get artifact by invalid reference
	req = httptest.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, "*&invalid-tag"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = handler.HeadManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	req = httptest.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName+"-none-exist", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = handler.HeadManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	req = httptest.NewRequest(http.MethodHead, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName+"-none-exist"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = handler.HeadManifest(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
