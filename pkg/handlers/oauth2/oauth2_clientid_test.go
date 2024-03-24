// Copyright 2024 sigma
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

package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestGetVersion(t *testing.T) {
	e := echo.New()
	validators.Initialize(e)
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oauth2Handler, err := handlerNew(inject{
		config: &configs.Configuration{
			Auth: configs.ConfigurationAuth{
				Jwt: configs.ConfigurationAuthJwt{
					PrivateKey: privateKeyString,
				},
				Oauth2: configs.ConfigurationAuthOauth2{
					Github: configs.ConfigurationAuthOauth2Github{
						ClientID: "github_client_id",
					},
					Gitlab: configs.ConfigurationAuthOauth2Gitlab{
						ClientID: "gitlab_client_id",
					},
					Gitea: configs.ConfigurationAuthOauth2Gitea{
						ClientID: "gitea_client_id",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("provider")
		c.SetParamValues(enums.ProviderGithub.String())
		err = oauth2Handler.ClientID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		response := rec.Body.Bytes()
		assert.Equal(t, "github_client_id", gjson.GetBytes(response, "client_id").String())
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("provider")
		c.SetParamValues(enums.ProviderGitlab.String())
		err = oauth2Handler.ClientID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		response := rec.Body.Bytes()
		assert.Equal(t, "gitlab_client_id", gjson.GetBytes(response, "client_id").String())
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("provider")
		c.SetParamValues(enums.ProviderGitea.String())
		err = oauth2Handler.ClientID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		response := rec.Body.Bytes()
		assert.Equal(t, "gitea_client_id", gjson.GetBytes(response, "client_id").String())
	}
}
