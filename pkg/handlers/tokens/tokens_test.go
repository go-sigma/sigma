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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/token"
	mockToken "github.com/go-sigma/sigma/pkg/utils/token/mocks"
)

func TestToken(t *testing.T) {
	logger.SetLevel("debug")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digCon := dig.New()
	err := digCon.Provide(func() *configs.Configuration {
		return &configs.Configuration{
			Auth: configs.ConfigurationAuth{
				Jwt: configs.ConfigurationAuthJwt{
					PrivateKey: privateKeyString,
				},
			},
		}
	})
	require.NoError(t, err)

	tokenStr := "mock-token-string"
	err = digCon.Provide(func() token.Service {
		tokenSvc := mockToken.NewMockService(ctrl)
		tokenSvc.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(id int64, expire time.Duration) (string, error) {
			return tokenStr, nil
		})
		return tokenSvc
	})
	require.NoError(t, err)

	e := tests.NewEcho()

	require.NoError(t, digCon.Provide(func() *echo.Echo { return e }))

	handler, err := handlerNew(digCon)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(consts.ContextUser, &models.User{ID: 1, Username: "test"})
	err = handler.Token(c)
	require.NoError(t, err)
	require.NotNil(t, rec.Body)
	require.Equal(t, tokenStr, gjson.Get(rec.Body.String(), "token").String())
}
