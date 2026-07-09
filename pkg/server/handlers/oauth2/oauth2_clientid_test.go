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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestClientID(t *testing.T) {
	digCon := dig.New()
	cfg := config.Configuration{
		Auth: config.ConfigurationAuth{
			Oauth2: config.ConfigurationAuthOauth2{
				Github: config.ConfigurationAuthOauth2Github{
					ClientID: "github_client_id",
				},
				Gitlab: config.ConfigurationAuthOauth2Gitlab{
					ClientID: "gitlab_client_id",
				},
				Gitea: config.ConfigurationAuthOauth2Gitea{
					ClientID: "gitea_client_id",
				},
			},
		},
	}
	err := digCon.Provide(func() *config.Configuration {
		return &cfg
	})
	require.NoError(t, err)
	require.NoError(t, digCon.Provide(func() token.Service { return nil }))
	require.NoError(t, digCon.Provide(func() repouser.UserRepository { return nil }))

	e := testkit.NewGin()
	require.NoError(t, digCon.Provide(func() *gin.Engine { return e }))
	require.NoError(t, validators.Initialize())

	handler := &handler{
		Config: &cfg,
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = req
		c.Params = gin.Params{{Key: "provider", Value: enums.ProviderGithub.String()}}
		handler.ClientID(c, &api.Oauth2ClientIDRequest{Provider: enums.ProviderGithub})
		assert.Equal(t, http.StatusOK, rec.Code)
		response := rec.Body.Bytes()
		assert.Equal(t, "github_client_id", gjson.GetBytes(response, "client_id").String())
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = req
		c.Params = gin.Params{{Key: "provider", Value: enums.ProviderGitlab.String()}}
		handler.ClientID(c, &api.Oauth2ClientIDRequest{Provider: enums.ProviderGitlab})
		assert.Equal(t, http.StatusOK, rec.Code)
		response := rec.Body.Bytes()
		assert.Equal(t, "gitlab_client_id", gjson.GetBytes(response, "client_id").String())
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = req
		c.Params = gin.Params{{Key: "provider", Value: enums.ProviderGitea.String()}}
		handler.ClientID(c, &api.Oauth2ClientIDRequest{Provider: enums.ProviderGitea})
		assert.Equal(t, http.StatusOK, rec.Code)
		response := rec.Body.Bytes()
		assert.Equal(t, "gitea_client_id", gjson.GetBytes(response, "client_id").String())
	}
}
