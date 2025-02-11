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

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestBasicAuthToken(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path { // nolint: gocritic
		case "/v2/":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.Header().Set("Www-Authenticate", `Basic realm="basic-realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  srv.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.NoError(t, err)
}

func TestTLSBasicAuthToken(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path { // nolint: gocritic
		case "/v2/":
			username, password, ok := r.BasicAuth()
			if ok {
				if username == cUsername && password == cPassword {
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.Header().Set("Www-Authenticate", `Basic realm="basic-realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  srv.URL,
			TlsVerify: false,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.NoError(t, err)
}

func TestBearerAuthToken(t *testing.T) {
	var wwwAuthenticate string

	cUsername := "sigma"
	cPassword := "sigma"
	token := "sigma"
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
		}
	}))
	defer srv.Close()

	wwwAuthenticate = fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, srv.URL+"/user/token", "registry.docker.io", "repository:library/alpine:pull")

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  srv.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.NoError(t, err)
}

func TestDoRequest(t *testing.T) {
	var wwwAuthenticate string

	cUsername := "sigma"
	cPassword := "sigma"
	token := "sigma"
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

	wwwAuthenticate = fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, srv.URL+"/user/token", "registry.docker.io", "repository:library/alpine:pull")

	f := NewClientsFactory()
	clients, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  srv.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.NoError(t, err)

	statusCode, _, bodyReader, err := clients.DoRequest(context.Background(), http.MethodGet, "/v2/_catalog", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	bodyBytes, err := io.ReadAll(bodyReader)
	assert.NoError(t, err)
	assert.Equal(t, `{"repositories":["library/alpine"]}`, strings.TrimSpace(string(bodyBytes)))
}

func TestDoRequestPing1(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  "http://127.0.0.1:10010",
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.Error(t, err)
}

func TestDoRequestPing2(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.Error(t, err)
}

func TestDoRequestPing3(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.Error(t, err)
}

func TestDoRequestPing4(t *testing.T) {
	cUsername := "sigma"
	cPassword := "sigma"

	wwwAuthenticate1 := fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, "http://127.0.0.1:3000/user/token", "registry.docker.io", "repository:library/alpine:pull")
	wwwAuthenticate2 := fmt.Sprintf(`Bearer realm="%s",service="%s",scope="%s"`, "http://127.0.0.1:3001/user/token", "registry.docker.io", "repository:library/alpine:pull")

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(echo.HeaderWWWAuthenticate, wwwAuthenticate1)
		w.Header().Add(echo.HeaderWWWAuthenticate, wwwAuthenticate2)
		w.WriteHeader(http.StatusUnauthorized)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.Error(t, err)
}

func TestDoRequestNoUsername(t *testing.T) {
	cPassword := "sigma"

	wwwAuthenticate := fmt.Sprintf(`Basic realm="%s",service="%s",scope="%s"`, "http://127.0.0.1:3000/user/token", "registry.docker.io", "repository:library/alpine:pull")

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add(echo.HeaderWWWAuthenticate, wwwAuthenticate)
		w.WriteHeader(http.StatusUnauthorized)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
			Username:  "",
			Password:  cPassword,
		}})
	assert.Error(t, err)
}

func TestDoRequestPingBasicAuth(t *testing.T) {
	cPassword := "sigma"
	cUsername := "sigma"

	var wwwAuthenticate string

	authTimes := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		authTimes++
		if authTimes == 1 {
			w.Header().Add(echo.HeaderWWWAuthenticate, wwwAuthenticate)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	wwwAuthenticate = fmt.Sprintf(`Basic realm="%s",service="%s",scope="%s"`, s.URL+"/token", "registry.docker.io", "repository:library/alpine:pull")

	f := NewClientsFactory()
	_, err := f.New(configs.Configuration{
		Log: configs.ConfigurationLog{
			ProxyLevel: enums.LogLevelDebug,
		},
		Proxy: configs.ConfigurationProxy{
			Endpoint:  s.URL,
			TlsVerify: true,
			Username:  cUsername,
			Password:  cPassword,
		}})
	assert.Error(t, err)
}
