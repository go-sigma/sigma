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

package artifact

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"testing"
// 	"time"

// 	"github.com/labstack/echo/v4"
// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	daomock "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// 	"github.com/go-sigma/sigma/pkg/server/validators"
// )

// func TestDeleteArtifact(t *testing.T) {
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize(e)
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	ctx := log.Logger.WithContext(context.Background())

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 	)

// 	userServiceFactory := dao.NewUserServiceFactory()
// 	userService := userServiceFactory.New()
// 	userObj := &models.User{Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 	assert.NoError(t, userService.Create(ctx, userObj))
// 	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
// 	namespaceService := namespaceServiceFactory.New()
// 	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))
// 	log.Info().Interface("namespace", namespaceObj).Msg("namespace created")
// 	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
// 	repositoryService := repositoryServiceFactory.New()
// 	repositoryObj := &models.Repository{Name: repositoryName, NamespaceID: namespaceObj.ID}
// 	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))
// 	artifactServiceFactory := dao.NewArtifactServiceFactory()
// 	artifactService := artifactServiceFactory.New()
// 	artifactObj := &models.Artifact{
// 		NamespaceID:  namespaceObj.ID,
// 		RepositoryID: repositoryObj.ID,
// 		Digest:       "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf",
// 		Size:         1234,
// 		ContentType:  "application/octet-stream",
// 		Raw:          []byte("test"),
// 		Blobs: []*models.Blob{
// 			{Digest: "sha256:123", Size: 123, ContentType: "test"},
// 			{Digest: "sha256:234", Size: 234, ContentType: "test"},
// 		},
// 	}
// 	assert.NoError(t, artifactService.Create(ctx, artifactObj))

// 	time.Sleep(time.Second * 3)

// 	tagServiceFactory := dao.NewTagServiceFactory()
// 	tagService := tagServiceFactory.New()
// 	tagObj := &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now().UnixMilli()}
// 	assert.NoError(t, tagService.Create(ctx, tagObj))

// 	artifactHandler := handlerNew()

// 	q := make(url.Values)
// 	q.Set("repository", "test/busybox")
// 	req := httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "digest")
// 	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf")
// 	assert.NoError(t, artifactHandler.DeleteArtifact(c))
// 	assert.Equal(t, http.StatusNoContent, c.Response().Status)

// 	q = make(url.Values)
// 	q.Set("repository", "test/busybox")
// 	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "digest")
// 	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5")
// 	assert.NoError(t, artifactHandler.DeleteArtifact(c))
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	q = make(url.Values)
// 	q.Set("repository", "test/busybox")
// 	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "digest")
// 	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf")
// 	assert.NoError(t, artifactHandler.DeleteArtifact(c))
// 	assert.Equal(t, http.StatusNotFound, c.Response().Status)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockArtifactService := daomock.NewMockArtifactService(ctrl)
// 	daoMockArtifactService.EXPECT().DeleteByDigest(gomock.Any(), gomock.Any(), gomock.Any()).
// 		DoAndReturn(func(_ context.Context, _, _ string) error {
// 			return fmt.Errorf("test")
// 		}).Times(1)
// 	daoMockArtifactServiceFactory := daomock.NewMockArtifactServiceFactory(ctrl)
// 	daoMockArtifactServiceFactory.EXPECT().New(gomock.Any()).
// 		DoAndReturn(func(txs ...*query.Query) dao.ArtifactService {
// 			return daoMockArtifactService
// 		}).Times(1)

// 	artifactHandler = handlerNew(inject{artifactServiceFactory: daoMockArtifactServiceFactory})

// 	q = make(url.Values)
// 	q.Set("repository", "test/busybox")
// 	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "digest")
// 	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf")
// 	assert.NoError(t, artifactHandler.DeleteArtifact(c))
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
