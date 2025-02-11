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
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/tidwall/gjson"
// 	"go.uber.org/mock/gomock"
// 	"gorm.io/gorm"

// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	daomocks "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
// 	workqmocks "github.com/go-sigma/sigma/pkg/modules/workq/definition/mocks"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// 	"github.com/go-sigma/sigma/pkg/server/validators"
// )

// func TestPostNamespace(t *testing.T) {
// 	logger.SetLevel("debug")

// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	e := echo.New()
// 	validators.Initialize(e)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	workQueueProducer := workqmocks.NewMockWorkQueueProducer(ctrl)
// 	workQueueProducer.EXPECT().Produce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, topic enums.Daemon, payload any, option definition.ProducerOption) error {
// 		return nil
// 	}).Times(3)

// 	namespaceHandler := handlerNew(inject{producerClient: workQueueProducer})

// 	userServiceFactory := dao.NewUserServiceFactory()
// 	userService := userServiceFactory.New()

// 	ctx := context.Background()
// 	userObj := &models.User{Username: "post-namespace", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
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
// 	resultID := gjson.GetBytes(rec.Body.Bytes(), "id").Int()
// 	assert.NotEqual(t, resultID, 0)

// 	// create with conflict
// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusConflict, c.Response().Status)

// 	// create with quota
// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test-quota","limit": 10000}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusCreated, c.Response().Status)

// 	// create with visibility
// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test-visibility","visibility": "private"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusCreated, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, "test")
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"t"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	daoMockNamespaceService := daomocks.NewMockNamespaceService(ctrl)
// 	daoMockNamespaceService.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.Namespace) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)

// 	daoMockNamespaceService.EXPECT().GetByName(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.Namespace, error) {
// 		return nil, gorm.ErrRecordNotFound
// 	}).Times(1)

// 	daoMockNamespaceServiceFactory := daomocks.NewMockNamespaceServiceFactory(ctrl)
// 	daoMockNamespaceServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.NamespaceService {
// 		return daoMockNamespaceService
// 	}).Times(2)

// 	namespaceHandler = handlerNew(inject{namespaceServiceFactory: daoMockNamespaceServiceFactory, producerClient: workQueueProducer})

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }

// // TestPostNamespaceCreateFailed1 get name failed
// func TestPostNamespaceCreateFailed1(t *testing.T) {
// 	logger.SetLevel("debug")

// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	e := echo.New()
// 	validators.Initialize(e)

// 	userServiceFactory := dao.NewUserServiceFactory()
// 	userService := userServiceFactory.New()

// 	ctx := context.Background()
// 	userObj := &models.User{Username: "post-namespace", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 	err := userService.Create(ctx, userObj)
// 	assert.NoError(t, err)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockNamespaceService := daomocks.NewMockNamespaceService(ctrl)
// 	daoMockNamespaceService.EXPECT().GetByName(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.Namespace, error) {
// 		return nil, fmt.Errorf("test")
// 	}).Times(1)

// 	daoMockNamespaceServiceFactory := daomocks.NewMockNamespaceServiceFactory(ctrl)
// 	daoMockNamespaceServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.NamespaceService {
// 		return daoMockNamespaceService
// 	}).Times(1)

// 	namespaceHandler := handlerNew(inject{namespaceServiceFactory: daoMockNamespaceServiceFactory})

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.PostNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
