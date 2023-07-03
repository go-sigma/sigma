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

package artifact

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/validators"
)

func TestListArtifact(t *testing.T) {
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
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)
		namespaceServiceFactory := dao.NewNamespaceServiceFactory()
		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: namespaceName, UserID: userObj.ID, Visibility: ptr.Of(enums.VisibilityPrivate)}
		err := namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)
		log.Info().Interface("namespace", namespaceObj).Msg("namespace created")
		repositoryServiceFactory := dao.NewRepositoryServiceFactory()
		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: namespaceName + "/" + repositoryName, NamespaceID: namespaceObj.ID, Visibility: ptr.Of(enums.VisibilityPrivate)}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)
		artifactServiceFactory := dao.NewArtifactServiceFactory()
		artifactService := artifactServiceFactory.New(tx)
		artifactObj := &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:e032eb458559f05c333b90abdeeac8ccb23bc1613137eeab2bbc0ea1224c5faf",
			Size:         1234,
			ContentType:  "application/octet-stream",
			Raw:          []byte("test"),
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
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req := httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = artifactHandler.ListArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, int64(1), gjson.GetBytes(rec.Body.Bytes(), "total").Int())

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = artifactHandler.ListArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockArtifactService := daomock.NewMockArtifactService(ctrl)
	daoMockArtifactServiceTimes := 0
	daoMockArtifactService.EXPECT().ListArtifact(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListArtifactRequest) ([]*models.Artifact, error) {
		daoMockArtifactServiceTimes++
		if daoMockArtifactServiceTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return []*models.Artifact{}, nil
	}).Times(3)
	daoMockArtifactService.EXPECT().CountArtifact(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListArtifactRequest) (int64, error) {
		return 0, fmt.Errorf("test")
	}).Times(1)
	daoMockArtifactServiceFactory := daomock.NewMockArtifactServiceFactory(ctrl)
	daoMockArtifactServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.ArtifactService {
		return daoMockArtifactService
	}).Times(3)

	daoMockTagService := daomock.NewMockTagService(ctrl)
	daoMockTagServiceTimes := 0
	daoMockTagService.EXPECT().CountByArtifact(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ []int64) (map[int64]int64, error) {
		daoMockTagServiceTimes++
		if daoMockTagServiceTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return map[int64]int64{1: 1}, nil
	}).Times(2)
	daoMockTagServiceFactory := daomock.NewMockTagServiceFactory(ctrl)
	daoMockTagServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.TagService {
		return daoMockTagService
	}).Times(2)

	artifactHandler = handlerNew(inject{artifactServiceFactory: daoMockArtifactServiceFactory, tagServiceFactory: daoMockTagServiceFactory})

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = artifactHandler.ListArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = artifactHandler.ListArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	q = make(url.Values)
	q.Set("repository", "test/busybox")
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodDelete, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = artifactHandler.ListArtifact(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
