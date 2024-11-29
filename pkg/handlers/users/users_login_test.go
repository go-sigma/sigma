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
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/labstack/echo/v4"
// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/go-sigma/sigma/pkg/configs"
// 	"github.com/go-sigma/sigma/pkg/consts"
// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/inits"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// )

// func TestLogin(t *testing.T) {
// 	logger.SetLevel("debug")

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
// 			Token: configs.ConfigurationAuthToken{
// 				Realm:   "http://localhost:8080/user/token",
// 				Service: "sigma-dev",
// 			},
// 		},
// 	}
// 	configs.SetConfiguration(config)

// 	assert.NoError(t, inits.Initialize(ptr.To(config)))

// 	userHandler, err := handlerNew()
// 	assert.NoError(t, err)

// 	userService := dao.NewUserServiceFactory().New()
// 	userObj, err := userService.GetByUsername(log.Logger.WithContext(context.Background()), "sigma")
// 	assert.NoError(t, err)

// 	e := echo.New()

// 	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"username":"sigma","password":"sigma"}`))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	c.Set(consts.ContextUser, userObj)
// 	err = userHandler.Login(c)
// 	assert.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, c.Response().Status)
// }
