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

package users

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/labstack/echo/v4"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/configs"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	daomock "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/dal/query"
// 	"github.com/go-sigma/sigma/pkg/inits"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	passwordmock "github.com/go-sigma/sigma/pkg/utils/password/mocks"
// 	tokenmock "github.com/go-sigma/sigma/pkg/utils/token/mocks"
// 	"github.com/go-sigma/sigma/pkg/server/validators"
// )

// func TestSignup(t *testing.T) {
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

// 	config := &configs.Configuration{
// 		Auth: configs.ConfigurationAuth{
// 			Admin: configs.ConfigurationAuthAdmin{
// 				Username: "sigma",
// 				Password: "sigma",
// 				Email:    "sigma@gmail.com",
// 			},
// 			Jwt: configs.ConfigurationAuthJwt{
// 				PrivateKey: privateKeyString,
// 			},
// 		},
// 	}
// 	configs.SetConfiguration(config)

// 	userHandler, err := handlerNew()
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusConflict, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"","password":"123Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

// 	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!123Aa!1","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec = httptest.NewRecorder()
// 	c = e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, c.Response().Status)
// }

// func TestSignupMockToken1(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	tokenMock := tokenmock.NewMockTokenService(ctrl)
// 	tokenMock.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(_ int64, _ time.Duration) (string, error) {
// 		return "test", nil
// 	}).Times(2)

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

// 	userHandler, err := handlerNew(inject{tokenService: tokenMock})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, c.Response().Status)
// }

// func TestSignupMockToken2(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	tokenMock := tokenmock.NewMockTokenService(ctrl)
// 	tokenMock.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(_ int64, _ time.Duration) (string, error) {
// 		return "", fmt.Errorf("test")
// 	}).Times(1)

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

// 	userHandler, err := handlerNew(inject{tokenService: tokenMock})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }

// func TestSignupMockToken3(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	var times int
// 	tokenMock := tokenmock.NewMockTokenService(ctrl)
// 	tokenMock.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(_ int64, _ time.Duration) (string, error) {
// 		if times == 0 {
// 			times++
// 			return "test", nil
// 		} else {
// 			return "", fmt.Errorf("test")
// 		}
// 	}).Times(2)

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

// 	userHandler, err := handlerNew(inject{tokenService: tokenMock})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }

// func TestSignupMockPassword(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	passwordMock := passwordmock.NewMockPassword(ctrl)
// 	passwordMock.EXPECT().Hash(gomock.Any()).DoAndReturn(func(_ string) (string, error) {
// 		return "", fmt.Errorf("test")
// 	}).Times(1)

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

// 	userHandler, err := handlerNew(inject{passwordService: passwordMock})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }

// func TestSignupMockDAO(t *testing.T) {
// 	logger.SetLevel("debug")

// 	e := echo.New()
// 	e.HideBanner = true
// 	e.HidePort = true
// 	validators.Initialize(e)
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	err := inits.Initialize(configs.Configuration{
// 		Auth: configs.ConfigurationAuth{
// 			Admin: configs.ConfigurationAuthAdmin{
// 				Username: "sigma",
// 				Password: "sigma",
// 				Email:    "sigma@gmail.com",
// 			},
// 			Jwt: configs.ConfigurationAuthJwt{
// 				PrivateKey: privateKeyString,
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	daoMockUserService := daomock.NewMockUserService(ctrl)
// 	daoMockUserService.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.User, error) {
// 		return nil, fmt.Errorf("test")
// 	}).Times(1)
// 	daoMockUserService.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.User) error {
// 		return fmt.Errorf("test")
// 	}).Times(1)

// 	daoMockUserServiceFactory := daomock.NewMockUserServiceFactory(ctrl)
// 	daoMockUserServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.UserService {
// 		return daoMockUserService
// 	}).Times(1)

// 	userHandler, err := handlerNew(inject{userServiceFactory: daoMockUserServiceFactory})
// 	assert.NoError(t, err)

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"test","password":"123498712311Aa!","email":"test@xx.com"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	err = userHandler.Signup(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
// }
