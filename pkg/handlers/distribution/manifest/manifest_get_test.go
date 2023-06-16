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

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
)

func TestParseRef(t *testing.T) {
	h := &handler{}
	refs := h.parseRef("latest")
	assert.Equal(t, refs.Tag, "latest")

	refs = h.parseRef("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
	assert.Equal(t, refs.Digest.String(), "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
}

func TestGetManifestFallbackProxy(t *testing.T) {
	logger.SetLevel("debug")
	viper.SetDefault("log.level", "debug")
	err := tests.Initialize()
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

	mux := http.NewServeMux()

	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc(fmt.Sprintf("/v2/test/busybox/manifests/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.manifest.v1+json")
		_, _ = w.Write([]byte(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796","size":800},"layers":[{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904","size":37869}],"annotations":{"category":"Infrastructure","licenses":"Apache-2.0","org.opencontainers.image.authors":"VMware, Inc.","org.opencontainers.image.description":"NGINX Open Source is a web server that can be also used as a reverse proxy, load balancer, and HTTP cache. Recommended for high-demanding sites due to its ability to provide faster content.","org.opencontainers.image.source":"https://github.com/bitnami/charts/tree/main/bitnami/nginx","org.opencontainers.image.title":"nginx","org.opencontainers.image.url":"https://bitnami.com","org.opencontainers.image.version":"15.0.2"}}`))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc(fmt.Sprintf("/v2/test/alpine/manifests/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.docker.distribution.manifest.list.v2+json")
		_, _ = w.Write([]byte(`{"manifests":[{"digest":"sha256:25fad2a32ad1f6f510e528448ae1ec69a28ef81916a004d3629874104f8a7f70","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"amd64","os":"linux"},"size":528},{"digest":"sha256:ae30c2911284159e0dc2f244b5e7a8b801b9c9f3449806d6e5591de22b65ce15","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm","os":"linux","variant":"v6"},"size":528},{"digest":"sha256:0b75b5bfd67c3ffaee0e951533407f6d45d53d7f4dd139fa0c09747b4849dd5d","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm","os":"linux","variant":"v7"},"size":528},{"digest":"sha256:e3bd82196e98898cae9fe7fbfd6e2436530485974dc4fb3b7ddb69134eda2407","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm64","os":"linux","variant":"v8"},"size":528},{"digest":"sha256:bd649691cf299c58fec56fb84a5067a915da6915897c6f846a6e317e5ff42a4d","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"386","os":"linux"},"size":528},{"digest":"sha256:8d42f68528a085fe2d936dcca64c642463744eb47312bb8e95863464550165ca","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"ppc64le","os":"linux"},"size":528},{"digest":"sha256:579fb3e58c23e1dba58ce7d06a14417954d0daaca4e28fa0358e941895d752f8","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"s390x","os":"linux"},"size":528}],"mediaType":"application\/vnd.docker.distribution.manifest.list.v2+json","schemaVersion":2}`))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc(fmt.Sprintf("/v2/test/alpine-invalid/manifests/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.manifest.v1+json")
		_, _ = w.Write([]byte(`{"manifests":[{"digest":"sha256:25fad2a32ad1f6f510e528448ae1ec69a28ef81916a004d3629874104f8a7f70","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"amd64","os":"linux"},"size":528},{"digest":"sha256:ae30c2911284159e0dc2f244b5e7a8b801b9c9f3449806d6e5591de22b65ce15","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm","os":"linux","variant":"v6"},"size":528},{"digest":"sha256:0b75b5bfd67c3ffaee0e951533407f6d45d53d7f4dd139fa0c09747b4849dd5d","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm","os":"linux","variant":"v7"},"size":528},{"digest":"sha256:e3bd82196e98898cae9fe7fbfd6e2436530485974dc4fb3b7ddb69134eda2407","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"arm64","os":"linux","variant":"v8"},"size":528},{"digest":"sha256:bd649691cf299c58fec56fb84a5067a915da6915897c6f846a6e317e5ff42a4d","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"386","os":"linux"},"size":528},{"digest":"sha256:8d42f68528a085fe2d936dcca64c642463744eb47312bb8e95863464550165ca","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"ppc64le","os":"linux"},"size":528},{"digest":"sha256:579fb3e58c23e1dba58ce7d06a14417954d0daaca4e28fa0358e941895d752f8","mediaType":"application\/vnd.docker.distribution.manifest.v2+json","platform":{"architecture":"s390x","os":"linux"},"size":528}],"mediaType":"application\/vnd.docker.distribution.manifest.list.v2+json","schemaVersion":2}`))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/library/busybox/manifests/sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(echo.HeaderContentType, "application/vnd.oci.image.manifest.v1+json")
		_, _ = w.Write([]byte(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796","size":800},"layers":[{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904","size":37869}],"annotations":{"category":"Infrastructure","licenses":"Apache-2.0","org.opencontainers.image.authors":"VMware, Inc.","org.opencontainers.image.description":"NGINX Open Source is a web server that can be also used as a reverse proxy, load balancer, and HTTP cache. Recommended for high-demanding sites due to its ability to provide faster content.","org.opencontainers.image.source":"https://github.com/bitnami/charts/tree/main/bitnami/nginx","org.opencontainers.image.title":"nginx","org.opencontainers.image.url":"https://bitnami.com","org.opencontainers.image.version":"15.0.2"}}`))
		w.WriteHeader(http.StatusOK)
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	viper.Reset()
	viper.SetDefault("log.level", "info")
	viper.SetDefault("proxy.endpoint", s.URL)
	viper.SetDefault("proxy.tlsVerify", true)

	h := &handler{
		proxyTaskServiceFactory: dao.NewProxyTaskServiceFactory(),
	}

	// test about application/vnd.oci.image.manifest.v1+json
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/test/busybox/manifests/%s", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Tag: tagName})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// test about application/vnd.docker.distribution.manifest.list.v2+json
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/test/alpine/manifests/%s", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Tag: tagName})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// test about invalid manifest content-type
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/test/alpine-invalid/manifests/%s", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Tag: tagName})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// test about artifact
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/library/busybox/manifests/%s", "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Digest: digest.Digest("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151")})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockProxyTaskService := daomock.NewMockProxyTaskService(ctrl)
	daoMockProxyTaskService.EXPECT().SaveProxyTaskTag(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.ProxyTaskTag) error {
		return fmt.Errorf("test")
	}).Times(1)

	daoMockProxyTaskService.EXPECT().SaveProxyTaskArtifact(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.ProxyTaskArtifact) error {
		return fmt.Errorf("test")
	}).Times(1)

	daoMockProxyTaskServiceFactory := daomock.NewMockProxyTaskServiceFactory(ctrl)
	daoMockProxyTaskServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.ProxyTaskService {
		return daoMockProxyTaskService
	}).Times(2)

	h = &handler{
		proxyTaskServiceFactory: daoMockProxyTaskServiceFactory,
	}

	// test about handler with a failed save proxy task
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/test/alpine/manifests/%s", tagName), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Tag: tagName})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// test about handler with a failed save proxy task
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v2/library/busybox/manifests/%s", "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151"), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = echo.New().NewContext(req, rec)
	err = h.getManifestFallbackProxy(c, repositoryName, Refs{Digest: digest.Digest("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0151")})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
