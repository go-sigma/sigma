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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestToken(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	validators.Initialize(e)
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	// viper.SetDefault("auth.internalUser.password", "internal-sigma")
	// viper.SetDefault("auth.internalUser.username", "internal-sigma")
	// viper.SetDefault("auth.admin.password", "sigma")
	// viper.SetDefault("auth.admin.username", "sigma")
	// viper.SetDefault("auth.jwt.privateKey", privateKeyString)

	viper.SetDefault("redis.url", "redis://"+miniredis.RunT(t).Addr())

	config := &configs.Configuration{
		Auth: configs.ConfigurationAuth{
			Admin: configs.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@gmail.com",
			},
			Jwt: configs.ConfigurationAuthJwt{
				PrivateKey: privateKeyString,
			},
		},
	}
	configs.SetConfiguration(config)
	assert.NoError(t, inits.Initialize(ptr.To(configs.GetConfiguration())))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	userHandler, err := handlerNew()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = userHandler.Token(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)

	userObj := &models.User{Username: "test-token", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(consts.ContextUser, userObj)
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
}

func TestTokenMockDAO(t *testing.T) {
	viper.Reset()
	logger.SetLevel("debug")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	validators.Initialize(e)
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	viper.SetDefault("redis.url", "redis://"+miniredis.RunT(t).Addr())

	config := &configs.Configuration{
		Auth: configs.ConfigurationAuth{
			Admin: configs.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@gmail.com",
			},
			Jwt: configs.ConfigurationAuthJwt{
				PrivateKey: privateKeyString,
			},
		},
	}
	configs.SetConfiguration(config)

	assert.NoError(t, inits.Initialize(ptr.To(configs.GetConfiguration())))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	userHandler, err := handlerNew()
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
