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
// 	"strconv"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/tidwall/gjson"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/auth"
// 	authmocks "github.com/go-sigma/sigma/pkg/auth/mocks"
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
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// 	"github.com/go-sigma/sigma/pkg/server/validators"
// )

// func TestDeleteNamespace(t *testing.T) {
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
// 	}).Times(2)

// 	authService := authmocks.NewMockAuthService(ctrl)
// 	authService.EXPECT().Namespace(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(user models.User, namespaceID int64, auth enums.Auth) (bool, error) {
// 		return true, nil
// 	}).Times(3)

// 	authServiceFactory := authmocks.NewMockAuthServiceFactory(ctrl)
// 	authServiceFactory.EXPECT().New().DoAndReturn(func() auth.AuthService {
// 		return authService
// 	}).Times(3)

// 	namespaceHandler := handlerNew(inject{producerClient: workQueueProducer, authServiceFactory: authServiceFactory})

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
// 	bytes := rec.Body.Bytes()
// 	resultID := gjson.GetBytes(bytes, "id").Int()

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("id")
// 	c.SetParamValues(strconv.FormatInt(resultID, 10))
// 	err = namespaceHandler.DeleteNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNoContent, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = namespaceHandler.DeleteNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("id")
// 	c.SetParamValues(strconv.FormatInt(resultID, 10))
// 	err = namespaceHandler.DeleteNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusNotFound, c.Response().Status)

// 	daoMockNamespaceService := daomock.NewMockNamespaceService(ctrl)
// 	daoMockNamespaceService.EXPECT().DeleteByID(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockNamespaceService.EXPECT().Get(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ int64) (*models.Namespace, error) {
// 		return &models.Namespace{}, nil
// 	}).Times(1)

// 	daoMockNamespaceServiceFactory := daomock.NewMockNamespaceServiceFactory(ctrl)
// 	daoMockNamespaceServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.NamespaceService {
// 		return daoMockNamespaceService
// 	}).Times(2)

// 	namespaceHandler = handlerNew(inject{namespaceServiceFactory: daoMockNamespaceServiceFactory, authServiceFactory: authServiceFactory})

// 	req = httptest.NewRequest(http.MethodDelete, "/", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	c.SetParamNames("id")
// 	c.SetParamValues(strconv.FormatInt(resultID, 10))
// 	err = namespaceHandler.DeleteNamespace(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
