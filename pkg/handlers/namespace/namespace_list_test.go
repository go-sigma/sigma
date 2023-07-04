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

package namespace

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/ximager/ximager/pkg/consts"
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

func TestListNamespace(t *testing.T) {
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

	namespaceHandler := handlerNew()

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()

	ctx := context.Background()
	userObj := &models.User{Provider: enums.ProviderLocal, Username: "list-namespace", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	err = userService.Create(ctx, userObj)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test","description":""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(consts.ContextUser, userObj)
	err = namespaceHandler.PostNamespace(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, c.Response().Status)

	q := make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = namespaceHandler.ListNamespace(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	assert.Equal(t, gjson.GetBytes(rec.Body.Bytes(), "total").Int(), int64(1))

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = namespaceHandler.ListNamespace(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var listNamespaceTimes int
	daoMockNamespaceService := daomock.NewMockNamespaceService(ctrl)
	daoMockNamespaceService.EXPECT().ListNamespace(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListNamespaceRequest) ([]*models.Namespace, error) {
		listNamespaceTimes++
		if listNamespaceTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return []*models.Namespace{}, nil
	}).Times(2)
	daoMockNamespaceService.EXPECT().CountNamespace(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ types.ListNamespaceRequest) (int64, error) {
		return 0, fmt.Errorf("test")
	}).Times(1)

	daoMockNamespaceServiceFactory := daomock.NewMockNamespaceServiceFactory(ctrl)
	daoMockNamespaceServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.NamespaceService {
		return daoMockNamespaceService
	}).Times(2)

	namespaceHandler = handlerNew(inject{namespaceServiceFactory: daoMockNamespaceServiceFactory})

	q = make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = namespaceHandler.ListNamespace(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)

	q = make(url.Values)
	q.Set("page_size", strconv.Itoa(100))
	q.Set("page_num", strconv.Itoa(1))
	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = namespaceHandler.ListNamespace(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
