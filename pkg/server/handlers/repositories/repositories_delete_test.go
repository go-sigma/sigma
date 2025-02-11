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

package repositories

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"testing"

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

// func TestDeleteRepository(t *testing.T) {
// 	logger.SetLevel("debug")
// 	e := echo.New()
// 	validators.Initialize(e)
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		err = tests.DB.DeInit()
// 		assert.NoError(t, err)
// 	}()

// 	repositoryFactory := dao.NewRepositoryServiceFactory()
// 	namespaceFactory := dao.NewNamespaceServiceFactory()

// 	const (
// 		namespaceName  = "test"
// 		repositoryName = "test/busybox"
// 	)

// 	var repoID int64

// 	var namespaceObj *models.Namespace
// 	err = query.Q.Transaction(func(tx *query.Query) error {
// 		ctx := log.Logger.WithContext(context.Background())

// 		userServiceFactory := dao.NewUserServiceFactory()
// 		userService := userServiceFactory.New(tx)
// 		userObj := &models.User{Username: "new-runner", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 		err = userService.Create(ctx, userObj)
// 		assert.NoError(t, err)
// 		namespaceService := namespaceFactory.New(tx)
// 		namespaceObj = &models.Namespace{Name: namespaceName, Visibility: enums.VisibilityPrivate}
// 		err := namespaceService.Create(ctx, namespaceObj)
// 		if err != nil {
// 			return err
// 		}

// 		repositoryService := repositoryFactory.New(tx)
// 		repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName, Visibility: enums.VisibilityPrivate}
// 		err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID})
// 		if err != nil {
// 			return err
// 		}

// 		repoID = repositoryObj.ID

// 		return nil
// 	})
// 	assert.NoError(t, err)

// 	repositoryHandler := handlerNew()

// 	req := httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "id")
// 	c.SetParamValues(namespaceName, strconv.FormatInt(repoID, 10))
// 	err = repositoryHandler.DeleteRepository(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNoContent, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = repositoryHandler.DeleteRepository(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "id")
// 	c.SetParamValues(namespaceName, strconv.FormatInt(repoID, 10))
// 	err = repositoryHandler.DeleteRepository(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNotFound, c.Response().Status)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockRepositoryService := daomock.NewMockRepositoryService(ctrl)
// 	daoMockRepositoryService.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockRepositoryService.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) (*models.Repository, error) {
// 		return &models.Repository{NamespaceID: namespaceObj.ID}, nil
// 	}).Times(1)

// 	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)
// 	daoMockRepositoryServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.RepositoryService {
// 		return daoMockRepositoryService
// 	}).Times(1)

// 	repositoryHandler = handlerNew(inject{repositoryServiceFactory: daoMockRepositoryServiceFactory})
// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.SetParamNames("namespace", "id")
// 	c.SetParamValues(namespaceName, strconv.FormatInt(repoID, 10))
// 	err = repositoryHandler.DeleteRepository(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
