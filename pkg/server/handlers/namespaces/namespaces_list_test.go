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

package namespaces

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strconv"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/tidwall/gjson"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	daomock "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
// 	workqmocks "github.com/go-sigma/sigma/pkg/modules/workq/definition/mocks"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// 	"github.com/go-sigma/sigma/pkg/server/validators"
// )

// func TestListNamespace(t *testing.T) {
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

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	workQueueProducer := workqmocks.NewMockWorkQueueProducer(ctrl)
// 	workQueueProducer.EXPECT().Produce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, topic enums.Daemon, payload any, option definition.ProducerOption) error {
// 		return nil
// 	}).Times(1)

// 	namespaceHandler := handlerNew(inject{producerClient: workQueueProducer})

// 	userServiceFactory := dao.NewUserServiceFactory()
// 	userService := userServiceFactory.New()

// 	ctx := context.Background()
// 	userObj := &models.User{Username: "list-namespace", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 	err := userService.Create(ctx, userObj)
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test","description":""}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusCreated, c.Response().Status)

// 	q := make(url.Values)
// 	q.Set("limit", strconv.Itoa(100))
// 	q.Set("last", strconv.Itoa(0))
// 	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = namespaceHandler.ListNamespaces(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, c.Response().Status)
// 	assert.Equal(t, int64(2), gjson.GetBytes(rec.Body.Bytes(), "total").Int())

// 	var listNamespaceTimes int
// 	daoMockNamespaceService := daomock.NewMockNamespaceService(ctrl)
// 	daoMockNamespaceService.EXPECT().ListNamespaceWithAuth(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64, _ *string, _ types.Pagination, _ types.Sortable) ([]*models.Namespace, int64, error) {
// 		listNamespaceTimes++
// 		if listNamespaceTimes == 1 {
// 			return nil, 0, fmt.Errorf("test")
// 		}
// 		return []*models.Namespace{}, 0, nil
// 	}).Times(1)

// 	daoMockNamespaceServiceFactory := daomock.NewMockNamespaceServiceFactory(ctrl)
// 	daoMockNamespaceServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.NamespaceService {
// 		return daoMockNamespaceService
// 	}).Times(1)

// 	namespaceHandler = handlerNew(inject{namespaceServiceFactory: daoMockNamespaceServiceFactory})

// 	q = make(url.Values)
// 	q.Set("page_size", strconv.Itoa(100))
// 	q.Set("page_num", strconv.Itoa(1))
// 	req = httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = namespaceHandler.ListNamespaces(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
