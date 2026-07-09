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

package tags

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/labstack/echo/v4"
// 	"log/slog"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/repository/registry"
// 	daomock "github.com/go-sigma/sigma/pkg/dal/repository/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/testkit"
// 	"github.com/go-sigma/sigma/pkg/api/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// 	"github.com/go-sigma/sigma/pkg/validators"
// )

// func TestDeleteTag(t *testing.T) {
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize()
// 	assert.NoError(t, testkit.Initialize(t))
// 	assert.NoError(t, testkit.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, testkit.DB.DeInit())
// 	}()

// 	ctx := context.Background()

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 	)

// 	userRepository := repouser.NewUserRepository()
// 	userRepository := userRepository.New()
// 	userObj := &models.User{Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com"), Role: enums.UserRoleAdmin}
// 	err := userRepository.Create(ctx, userObj)
// 	assert.NoError(t, err)
// 	namespaceRepository := reponamespace.NewNamespaceRepository()
// 	namespaceRepository := namespaceRepository.New()
// 	namespaceObj := &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 	err = namespaceRepository.Create(ctx, namespaceObj)
// 	assert.NoError(t, err)
// 	slog.Info("namespace created", "namespace", namespaceObj)
// 	repositoryRepository := reporegistry.NewRepositoryRepository()
// 	repositoryRepository := repositoryRepository.New()
// 	repositoryObj := &models.Repository{Name: repositoryName, NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
// 	err = repositoryRepository.Create(ctx, repositoryObj, reporegistry.AutoCreateNamespace{UserID: userObj.ID})
// 	assert.NoError(t, err)
// 	artifactRepository := reporegistry.NewArtifactRepository()
// 	artifactRepository := artifactRepository.New()
// 	artifactObj := &models.Artifact{RepositoryID: repositoryObj.ID, Digest: "sha256:1234567890", Size: 1234, ContentType: "application/octet-stream", Raw: []byte("test"), PushedAt: time.Now()}
// 	err = artifactRepository.Create(ctx, artifactObj)
// 	assert.NoError(t, err)
// 	tagRepository := reporegistry.NewTagRepository()
// 	tagRepository := tagRepository.New()
// 	tagObj := &models.Tag{Name: "latest", RepositoryID: repositoryObj.ID, ArtifactID: artifactObj.ID, PushedAt: time.Now()}
// 	err = tagRepository.Create(ctx, tagObj)
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

// 	daoMockTagRepository := daomock.NewMockTagRepository(ctrl)
// 	daoMockTagRepository.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockTagRepository := daomock.NewMockTagRepository(ctrl)
// 	daoMockTagRepository.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) reporegistry.TagRepository {
// 		return daoMockTagRepository
// 	}).Times(1)

// 	tagHandler = handlerNew(inject{tagRepository: daoMockTagRepository})

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
