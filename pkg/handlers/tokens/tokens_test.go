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

package token

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	daomock "github.com/go-sigma/sigma/pkg/dal/dao/mocks"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	passwordmock "github.com/go-sigma/sigma/pkg/utils/password/mocks"
	tokenmock "github.com/go-sigma/sigma/pkg/utils/token/mocks"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestToken(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
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

	viper.SetDefault("auth.internalUser.password", "internal-sigma")
	viper.SetDefault("auth.internalUser.username", "internal-sigma")
	viper.SetDefault("auth.admin.password", "sigma")
	viper.SetDefault("auth.admin.username", "sigma")
	viper.SetDefault("auth.jwt.privateKey", privateKeyString)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err = inits.Initialize()
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var tokenTimes int
	tokenMock := tokenmock.NewMockTokenService(ctrl)
	tokenMock.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(_ *models.User, _ time.Duration) (string, error) {
		if tokenTimes == 0 {
			tokenTimes++
			return "test", nil
		} else {
			return "", fmt.Errorf("test")
		}
	}).Times(2)
	var passwordTimes int
	passwordMock := passwordmock.NewMockPassword(ctrl)
	passwordMock.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ string) bool {
		if passwordTimes == 0 { // nolint: gocritic
			passwordTimes++
			return true
		} else if passwordTimes == 1 {
			passwordTimes++
			return false
		} else {
			return true
		}
	}).Times(3)

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	userHandler, err := handlerNew(inject{tokenService: tokenMock, passwordService: passwordMock})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}

func TestTokenMockDAO(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
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

	viper.SetDefault("auth.internalUser.password", "internal-sigma")
	viper.SetDefault("auth.internalUser.username", "internal-sigma")
	viper.SetDefault("auth.admin.password", "sigma")
	viper.SetDefault("auth.admin.username", "sigma")
	viper.SetDefault("auth.jwt.privateKey", privateKeyString)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err = inits.Initialize()
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockUserService := daomock.NewMockUserService(ctrl)
	daoMockUserService.EXPECT().GetByUsername(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) (*models.User, error) {
		return nil, fmt.Errorf("test")
	}).Times(1)
	daoMockUserServiceFactory := daomock.NewMockUserServiceFactory(ctrl)
	daoMockUserServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.UserService {
		return daoMockUserService
	}).Times(1)

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	userHandler, err := handlerNew(inject{userServiceFactory: daoMockUserServiceFactory})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
}
