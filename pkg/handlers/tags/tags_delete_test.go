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

package tag

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/validators"
)

func TestDeleteTag(t *testing.T) {
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

	ctx := context.Background()

	const (
		namespaceName  = "test"
		repositoryName = "test/busybox"
	)

	var tagObj *models.Tag
	err = query.Q.Transaction(func(tx *query.Query) error {
		userServiceFactory := dao.NewUserServiceFactory()
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
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
		repositoryObj := &models.Repository{Name: repositoryName, NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)
		artifactServiceFactory := dao.NewArtifactServiceFactory()
		artifactService := artifactServiceFactory.New(tx)
		artifactObj := &models.Artifact{RepositoryID: repositoryObj.ID, Digest: "sha256:1234567890", Size: 1234, ContentType: "application/octet-stream", Raw: []byte("test"), PushedAt: time.Now()}
		err = artifactService.Create(ctx, artifactObj)
		assert.NoError(t, err)
		tagServiceFactory := dao.NewTagServiceFactory()
		tagService := tagServiceFactory.New(tx)
		tagObj = &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now()}
		err = tagService.Create(ctx, tagObj)
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	tagHandler := handlerNew()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	q := req.URL.Query()
	q.Add("repository", repositoryName)
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("namespace", "id")
	c.SetParamValues(namespaceName, strconv.FormatInt(tagObj.ID, 10))
	err = tagHandler.DeleteTag(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	q = req.URL.Query()
	q.Add("repository", repositoryName)
	req.URL.RawQuery = q.Encode()
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatInt(tagObj.ID, 10))
	err = tagHandler.DeleteTag(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockTagService := daomock.NewMockTagService(ctrl)
	daoMockTagService.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) error {
		return fmt.Errorf("test")
	}).Times(1)
	daoMockTagServiceFactory := daomock.NewMockTagServiceFactory(ctrl)
	daoMockTagServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.TagService {
		return daoMockTagService
	}).Times(1)

	tagHandler = handlerNew(inject{tagServiceFactory: daoMockTagServiceFactory})

	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	q = req.URL.Query()
	q.Add("repository", repositoryName)
	req.URL.RawQuery = q.Encode()
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace", "id")
	c.SetParamValues(namespaceName, strconv.FormatInt(tagObj.ID, 10))
	err = tagHandler.DeleteTag(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
