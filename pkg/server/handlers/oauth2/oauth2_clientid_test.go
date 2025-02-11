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
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

func TestClientID(t *testing.T) {
	digCon := dig.New()
	err := digCon.Provide(func() configs.Configuration {
		return configs.Configuration{
			Auth: configs.ConfigurationAuth{
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
		}
	})
	require.NoError(t, err)
	require.NoError(t, digCon.Provide(func() token.Service { return nil }))
	require.NoError(t, digCon.Provide(func() dao.UserServiceFactory { return nil }))

	e := tests.NewEcho()
	require.NoError(t, digCon.Provide(func() *echo.Echo { return e }))
	require.NoError(t, validators.Initialize(digCon))

	handler := handlerNew(digCon)
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("provider")
		c.SetParamValues(enums.ProviderGithub.String())
		err := handler.ClientID(c)
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
		err := handler.ClientID(c)
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
		err := handler.ClientID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		response := rec.Body.Bytes()
		assert.Equal(t, "gitea_client_id", gjson.GetBytes(response, "client_id").String())
	}
}
