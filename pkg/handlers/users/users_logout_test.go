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

package user

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/logger"
	tokenmock "github.com/go-sigma/sigma/pkg/utils/token/mocks"
)

func TestLogout(t *testing.T) {
	logger.SetLevel("debug")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var times int
	tokenMock := tokenmock.NewMockTokenService(ctrl)
	tokenMock.EXPECT().Revoke(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) error {
		if times == 0 {
			times++
			return nil
		} else {
			return fmt.Errorf("error")
		}
	}).Times(2)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	userHandler, err := handlerNew(inject{tokenService: tokenMock})
	assert.NoError(t, err)

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(consts.ContextJti, "")
	err = userHandler.Logout(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(consts.ContextJti, "test")
	err = userHandler.Logout(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(consts.ContextJti, "test")
	err = userHandler.Logout(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
}
