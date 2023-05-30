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

package repository

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/validators"
)

func TestDeleteRepository(t *testing.T) {
	logger.SetLevel("debug")
	e := echo.New()
	validators.Initialize(e)
	err := tests.Initialize()
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

	repositoryFactory := dao.NewRepositoryServiceFactory()
	namespaceFactory := dao.NewNamespaceServiceFactory()

	const (
		namespaceName  = "test"
		repositoryName = "busybox"
	)

	var repoID uint64

	err = query.Q.Transaction(func(tx *query.Query) error {
		ctx := log.Logger.WithContext(context.Background())

		namespaceService := namespaceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: namespaceName}
		err := namespaceService.Create(ctx, namespaceObj)
		if err != nil {
			return err
		}

		repositoryService := repositoryFactory.New(tx)
		repositoryObj := &models.Repository{NamespaceID: namespaceObj.ID, Name: repositoryName}
		err = repositoryService.Create(ctx, repositoryObj)
		if err != nil {
			return err
		}

		repoID = repositoryObj.ID

		return nil
	})
	assert.NoError(t, err)

	repositoryHandler := handlerNew()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatUint(repoID, 10))
	err = repositoryHandler.DeleteRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, c.Response().Status)

	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = repositoryHandler.DeleteRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatUint(repoID, 10))
	err = repositoryHandler.DeleteRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, c.Response().Status)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockRepositoryService := daomock.NewMockRepositoryService(ctrl)
	daoMockRepositoryService.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ uint64) error {
		return fmt.Errorf("test")
	}).Times(1)

	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)
	daoMockRepositoryServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.RepositoryService {
		return daoMockRepositoryService
	}).Times(1)

	repositoryHandler = handlerNew(inject{repositoryServiceFactory: daoMockRepositoryServiceFactory})
	req = httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.FormatUint(repoID, 10))
	err = repositoryHandler.DeleteRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
