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

package token

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/inits"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	passwordmock "github.com/ximager/ximager/pkg/utils/password/mocks"
	tokenmock "github.com/ximager/ximager/pkg/utils/token/mocks"
	"github.com/ximager/ximager/pkg/validators"
)

func TestToken(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
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

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	viper.SetDefault("auth.admin.password", "ximager")
	viper.SetDefault("auth.admin.username", "ximager")
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
	req.SetBasicAuth("ximager", "ximager")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("ximager", "ximager")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("ximager", "ximager")
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

	viper.SetDefault("auth.internalUser.password", "internal-ximager")
	viper.SetDefault("auth.internalUser.username", "internal-ximager")
	viper.SetDefault("auth.admin.password", "ximager")
	viper.SetDefault("auth.admin.username", "ximager")
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
	req.SetBasicAuth("ximager", "ximager")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
}
