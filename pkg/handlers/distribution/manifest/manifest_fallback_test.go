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

package manifest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/types"
)

func TestFallbackProxy(t *testing.T) {
	var wwwAuthenticate string

	cUsername := "ximager"
	cPassword := "ximager"
	token := "ximager"
	service := "registry.docker.io"
	scope := "repository:library/alpine:pull"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			if r.Header.Get("Authorization") == "Bearer "+token {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Www-Authenticate", wwwAuthenticate)
			w.WriteHeader(http.StatusUnauthorized)
		case "/user/token":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					query := r.URL.Query()
					svc := query.Get("service")
					so := query.Get("scope")
					if svc != service || so != scope {
						t.Error("service or scope not match")
					}
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(types.PostUserTokenResponse{Token: token})
					assert.NoError(t, err)
					return
				}
			}
		case "/v2/_catalog":
			if r.Header.Get("Authorization") == "Bearer "+token {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(dtspecv1.RepositoryList{Repositories: []string{"library/alpine"}})
				assert.NoError(t, err)
				return
			}
			w.Header().Set("Www-Authenticate", wwwAuthenticate)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	viper.SetDefault("log.level", "info")
	viper.SetDefault("proxy.endpoint", srv.URL)
	viper.SetDefault("proxy.tlsVerify", true)
	viper.SetDefault("proxy.username", cUsername)
	viper.SetDefault("proxy.password", cPassword)

	wwwAuthenticate = fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, srv.URL+"/user/token", "registry.docker.io", "repository:library/alpine:pull")

	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	statusCode, _, bodyBytes, err := fallbackProxy(c)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, `{"repositories":["library/alpine"]}`, strings.TrimSpace(string(bodyBytes)))
}
