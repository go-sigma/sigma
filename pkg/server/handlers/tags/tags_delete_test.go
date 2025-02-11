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

package tag

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/labstack/echo/v4"
// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/consts"
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

// func TestDeleteTag(t *testing.T) {
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

// 	ctx := context.Background()

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 	)

// 	userServiceFactory := dao.NewUserServiceFactory()
// 	userService := userServiceFactory.New()
// 	userObj := &models.User{Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com"), Role: enums.UserRoleAdmin}
// 	err := userService.Create(ctx, userObj)
// 	assert.NoError(t, err)
// 	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
// 	namespaceService := namespaceServiceFactory.New()
// 	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 	err = namespaceService.Create(ctx, namespaceObj)
// 	assert.NoError(t, err)
// 	log.Info().Interface("namespace", namespaceObj).Msg("namespace created")
// 	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
// 	repositoryService := repositoryServiceFactory.New()
// 	repositoryObj := &models.Repository{Name: repositoryName, NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
// 	err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID})
// 	assert.NoError(t, err)
// 	artifactServiceFactory := dao.NewArtifactServiceFactory()
// 	artifactService := artifactServiceFactory.New()
// 	artifactObj := &models.Artifact{RepositoryID: repositoryObj.ID, Digest: "sha256:1234567890", Size: 1234, ContentType: "application/octet-stream", Raw: []byte("test"), PushedAt: time.Now()}
// 	err = artifactService.Create(ctx, artifactObj)
// 	assert.NoError(t, err)
// 	tagServiceFactory := dao.NewTagServiceFactory()
// 	tagService := tagServiceFactory.New()
// 	tagObj := &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now()}
// 	err = tagService.Create(ctx, tagObj)
// 	assert.NoError(t, err)

// 	tagHandler := handlerNew()

// 	req := httptest.NewRequest(http.MethodDelete, "/", nil)
// 	q := req.URL.Query()
// 	q.Add("repository", repositoryName)
// 	req.URL.RawQuery = q.Encode()
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("namespace", "id")
// 	c.SetParamValues(namespaceName, strconv.FormatInt(tagObj.ID, 10))
// 	err = tagHandler.DeleteTag(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNoContent, rec.Code)

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	q = req.URL.Query()
// 	q.Add("repository", repositoryName)
// 	req.URL.RawQuery = q.Encode()
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("id")
// 	c.SetParamValues(strconv.FormatInt(tagObj.ID, 10))
// 	err = tagHandler.DeleteTag(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, rec.Code)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockTagService := daomock.NewMockTagService(ctrl)
// 	daoMockTagService.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockTagServiceFactory := daomock.NewMockTagServiceFactory(ctrl)
// 	daoMockTagServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.TagService {
// 		return daoMockTagService
// 	}).Times(1)

// 	tagHandler = handlerNew(inject{tagServiceFactory: daoMockTagServiceFactory})

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	q = req.URL.Query()
// 	q.Add("repository", repositoryName)
// 	req.URL.RawQuery = q.Encode()
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("namespace", "id")
// 	c.SetParamValues(namespaceName, strconv.FormatInt(tagObj.ID, 10))
// 	err = tagHandler.DeleteTag(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
// }
