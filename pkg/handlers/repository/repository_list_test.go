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
	"net/url"
	"strconv"
	"testing"

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
	"github.com/ximager/ximager/pkg/validators"
)

func TestListRepository(t *testing.T) {
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

		return nil
	})
	assert.NoError(t, err)

	repositoryHandler := handlerNew()

	q := make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = repositoryHandler.ListRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, int64(1), gjson.GetBytes(rec.Body.Bytes(), "total").Int())

	q = make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = repositoryHandler.ListRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockRepositoryService := daomock.NewMockRepositoryService(ctrl)
	var listRepositoryTimes int
	daoMockRepositoryService.EXPECT().ListRepository(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListRepositoryRequest) ([]*models.Repository, error) {
		listRepositoryTimes++
		if listRepositoryTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return []*models.Repository{}, nil
	}).Times(2)
	daoMockRepositoryService.EXPECT().CountRepository(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListRepositoryRequest) (int64, error) {
		return 0, fmt.Errorf("test")
	}).Times(1)

	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)
	daoMockRepositoryServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.RepositoryService {
		return daoMockRepositoryService
	}).Times(2)

	repositoryHandler = handlerNew(inject{repositoryServiceFactory: daoMockRepositoryServiceFactory})
	q = make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = repositoryHandler.ListRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	q = make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("namespace")
	c.SetParamValues(namespaceName)
	err = repositoryHandler.ListRepository(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
