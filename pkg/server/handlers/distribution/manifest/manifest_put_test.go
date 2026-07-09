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
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/alicebob/miniredis/v2"
// 	"github.com/labstack/echo/v4"
// 	"log/slog"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/authz"
// 	"github.com/go-sigma/sigma/pkg/config"
// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/repository/registry"
// 	daomock "github.com/go-sigma/sigma/pkg/dal/repository/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/testkit"
// 	"github.com/go-sigma/sigma/pkg/api/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// )

// func TestPutManifestAsyncTask(t *testing.T) {
// 	assert.NoError(t, testkit.Initialize(t))
// 	assert.NoError(t, testkit.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, testkit.DB.DeInit())
// 	}()
// 	viper.SetDefault("redis.url", "redis://"+miniredis.RunT(t).Addr())

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockArtifactRepository := daomock.NewMockArtifactRepository(ctrl)
// 	daoMockArtifactRepository.EXPECT().CreateSbom(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.ArtifactSbom) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockArtifactRepository.EXPECT().CreateVulnerability(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.ArtifactVulnerability) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)

// 	daoMockArtifactRepository := daomock.NewMockArtifactRepository(ctrl)
// 	daoMockArtifactRepository.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) reporegistry.ArtifactRepository {
// 		return daoMockArtifactRepository
// 	}).Times(2)

// 	h := &handler{
// 		artifactRepository: daoMockArtifactRepository,
// 	}

// 	ctx := context.Background()
// 	h.putManifestAsyncTask(ctx, &models.Artifact{ID: 1})
// }

// func TestPutManifest(t *testing.T) {
// 	assert.NoError(t, testkit.Initialize(t))
// 	assert.NoError(t, testkit.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, testkit.DB.DeInit())
// 	}()

// 	viper.SetDefault("redis.url", "redis://"+miniredis.RunT(t).Addr())

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 		digestName     = "sha256:2776ee23722eaabcffed77dafd22b7a1da734971bf268a323b6819926dfe1ebd" // nolint: gosec
// 		tagName        = "latest"
// 	)

// 	ctx := context.Background()

// 	userObj := &models.User{Username: "head-manifest", Password: ptr.Of("test"), Role: enums.UserRoleRoot, Email: ptr.Of("test@gmail.com")}
// 	assert.NoError(t, repouser.NewUserRepository().Create(ctx, userObj))

// 	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, reponamespace.NewNamespaceRepository().Create(ctx, namespaceObj))

// 	repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName}
// 	assert.NoError(t, reporegistry.NewRepositoryRepository().Create(ctx, repositoryObj, reporegistry.AutoCreateNamespace{AutoCreate: false, Visibility: enums.VisibilityPrivate, UserID: userObj.ID}))

// 	blobRepository := reporegistry.NewBlobRepository()
// 	blobLayer1 := &models.Blob{Digest: "sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796", Size: 123, ContentType: "application/vnd.cncf.helm.config.v1+json"}
// 	assert.NoError(t, blobRepository.Create(ctx, blobLayer1))
// 	blobLayer2 := &models.Blob{Digest: "sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904", Size: 122, ContentType: "application/vnd.cncf.helm.chart.content.v1.tar+gzip"}
// 	assert.NoError(t, blobRepository.Create(ctx, blobLayer2))

// 	h := &handler{
// 		config: &config.Configuration{
// 			Namespace: config.ConfigurationNamespace{
// 				AutoCreate: true,
// 				Visibility: enums.VisibilityPublic,
// 			},
// 		},
// 		authService:       authz.NewAuthorizer(nil, nil, nil, nil, nil),
// 		namespaceRepository:  reponamespace.NewNamespaceRepository(),
// 		repositoryRepository: reporegistry.NewRepositoryRepository(),
// 		tagRepository:        reporegistry.NewTagRepository(),
// 		artifactRepository:   reporegistry.NewArtifactRepository(),
// 		blobRepository:       reporegistry.NewBlobRepository(),
// 	}

// 	// test about put manifest
// 	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", repositoryName, tagName), bytes.NewReader([]byte(`{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json1","digest":"sha256:a61fd63bebd559934a60e30d1e7b832a136ac6bae3a11ca97ade20bfb3645796","size":800},"layers":[{"mediaType":"application/vnd.cncf.helm.chart.content.v1.tar+gzip","digest":"sha256:e45dd3e880e94bdb52cc88d6b4e0fbaec6876856f39a1a89f76e64d0739c2904","size":37869}],"annotations":{"category":"Infrastructure","licenses":"Apache-2.0","org.opencontainers.image.authors":"VMware, Inc.","org.opencontainers.image.description":"NGINX Open Source is a web server that can be also used as a reverse proxy, load balancer, and HTTP cache. Recommended for high-demanding sites due to its ability to provide faster content.","org.opencontainers.image.source":"https://github.com/bitnami/charts/tree/main/bitnami/nginx","org.opencontainers.image.title":"nginx","org.opencontainers.image.url":"https://bitnami.com","org.opencontainers.image.version":"15.0.2"}}`)))
// 	req.Header.Set(echo.HeaderContentType, "application/vnd.oci.image.manifest.v1+json")
// 	rec := httptest.NewRecorder()
// 	c := echo.New().NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	assert.NoError(t, h.PutManifest(c))
// 	assert.Equal(t, http.StatusCreated, rec.Code)
// }
