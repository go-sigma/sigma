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

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	daomock "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestGetArtifact(t *testing.T) {
	logger.SetLevel("debug")
	e := echo.New()
	validators.Initialize(e)
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	ctx := log.Logger.WithContext(context.Background())

	const (
		namespaceName  = "test"
		repositoryName = "busybox"
	)

	err = query.Q.Transaction(func(tx *query.Query) error {
		userServiceFactory := dao.NewUserServiceFactory()
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)
		namespaceServiceFactory := dao.NewNamespaceServiceFactory()
		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
		err := namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)
		log.Info().Interface("namespace", namespaceObj).Msg("namespace created")
		repositoryServiceFactory := dao.NewRepositoryServiceFactory()
		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: namespaceName + "/" + repositoryName, NamespaceID: namespaceObj.ID}
		err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)
		artifactServiceFactory := dao.NewArtifactServiceFactory()
		artifactService := artifactServiceFactory.New(tx)
		artifactObj := &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf",
			Size:         1234,
			ContentType:  "application/octet-stream",
			Raw:          []byte("test"),
			PushedAt:     time.Now(),
			Blobs:        []*models.Blob{{Digest: "sha256:123", Size: 123, ContentType: "test"}, {Digest: "sha256:234", Size: 234, ContentType: "test"}},
		}
		err = artifactService.Create(ctx, artifactObj)
		assert.NoError(t, err)
		tagServiceFactory := dao.NewTagServiceFactory()
		tagService := tagServiceFactory.New(tx)
		tagObj := &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now()}
		err = tagService.Create(ctx, tagObj)
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	artifactHandler := handlerNew()

	q := make(url.Values)
	q.Set("repository", "test/busybox")
	req := httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("namespace", "digest")
	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf", gjson.GetBytes(rec.Body.Bytes(), "digest").String())

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("digest")
	c.SetParamValues("sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace", "digest")
	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5f1f")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	q = make(url.Values)
	q.Set("repository", "test/busybox-none-exist")
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace", "digest")
	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5f1f")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockArtifactService := daomock.NewMockArtifactService(ctrl)
	daoMockArtifactService.EXPECT().GetByDigest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64, _ string) (*models.Artifact, error) {
		return nil, fmt.Errorf("test")
	}).Times(1)
	daoMockArtifactServiceFactory := daomock.NewMockArtifactServiceFactory(ctrl)
	daoMockArtifactServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.ArtifactService {
		return daoMockArtifactService
	}).Times(1)

	artifactHandler = handlerNew(inject{artifactServiceFactory: daoMockArtifactServiceFactory})

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace", "digest")
	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5f1f")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	daoMockRepositoryService := daomock.NewMockRepositoryService(ctrl)
	daoMockRepositoryService.EXPECT().GetByName(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.Repository, error) {
		return nil, fmt.Errorf("test")
	}).Times(1)
	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)
	daoMockRepositoryServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.RepositoryService {
		return daoMockRepositoryService
	}).Times(1)

	artifactHandler = handlerNew(inject{repositoryServiceFactory: daoMockRepositoryServiceFactory})

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace", "digest")
	c.SetParamValues(namespaceName, "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5f1f")
	err = artifactHandler.GetArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
