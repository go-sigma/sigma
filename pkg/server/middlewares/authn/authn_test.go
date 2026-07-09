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

package authn

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/app/bootstrap"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"

	_ "github.com/go-sigma/sigma/pkg/infra/lock/inmemory"
)

var privateKeyString = func() string {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		panic(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	return base64.StdEncoding.EncodeToString(pemBytes)
}()

func TestGenWwwAuthenticate(t *testing.T) {
	logger.SetLevel("debug")

	config := config.Configuration{
		Auth: config.ConfigurationAuth{
			Admin: config.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@tosone.cn",
			},
			Jwt: config.ConfigurationAuthJwt{
				PrivateKey: privateKeyString,
			},
			Token: config.ConfigurationAuthToken{
				Realm:   "http://localhost:8080/user/token",
				Service: "sigma-dev",
			},
		},
	}

	type args struct {
		host   string
		schema string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test genWwwAuthenticate",
			args: args{
				host:   "localhost:8080",
				schema: "http",
			},
			want: "Bearer realm=\"http://localhost:8080/user/token\",service=\"sigma-dev\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genWwwAuthenticate(&config, tt.args.host, tt.args.schema); got != tt.want {
				t.Errorf("genWwwAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func genAuthConfig(t *testing.T, digCon *dig.Container, skipper Skipper) Config {
	var cfg *config.Configuration
	require.NoError(t, digCon.Invoke(func(c *config.Configuration) { cfg = c }))
	var passwordSvc password.Service
	require.NoError(t, digCon.Invoke(func(s password.Service) { passwordSvc = s }))
	var userRepository repouser.UserRepository
	require.NoError(t, digCon.Invoke(func(f repouser.UserRepository) { userRepository = f }))
	tokenSvc := genTokenService(t, digCon)
	return Config{
		Skipper:        skipper,
		Config:         cfg,
		TokenSvc:       tokenSvc,
		PasswordSvc:    passwordSvc,
		UserRepository: userRepository,
	}
}

func genTokenService(t *testing.T, digCon *dig.Container) token.Service {
	var cfg *config.Configuration
	require.NoError(t, digCon.Invoke(func(c *config.Configuration) { cfg = c }))
	var redisClientFactory dalredis.ClientFactory
	_ = digCon.Invoke(func(f dalredis.ClientFactory) { redisClientFactory = f })
	tokenSvc, err := token.New(token.Params{
		Config:             cfg,
		RedisClientFactory: redisClientFactory,
	})
	require.NoError(t, err)
	return tokenSvc
}

func TestAuthWithConfig(t *testing.T) {
	logger.SetLevel("debug")
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		genDigCon     func(*testing.T) *dig.Container
		genAuthConfig func(*testing.T, *dig.Container) Config
		afterCheck    func(*testing.T, *dig.Container, gin.HandlerFunc)
	}{
		{
			name: "normal",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() *config.Configuration {
					return &config.Configuration{
						Auth: config.ConfigurationAuth{
							Admin: config.ConfigurationAuthAdmin{
								Username: "sigma",
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
						Locker: config.ConfigurationLocker{
							Type:   enums.LockerTypeInmemory,
							Prefix: "sigma-locker",
						},
						Redis: config.ConfigurationRedis{},
						Cache: config.ConfigurationCache{
							Type:   enums.CacherTypeInmemory,
							Prefix: "sigma-cache",
							Inmemory: config.ConfigurationCacheInmemory{
								Size: 100,
							},
						},
					}
				})
				require.NoError(t, err)

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				require.NoError(t, testkit.InitializeIntegration(t, digCon))

				return digCon
			},
			genAuthConfig: func(t *testing.T, c *dig.Container) Config {
				return genAuthConfig(t, c, nil)
			},
			afterCheck: func(t *testing.T, digCon *dig.Container, middleware gin.HandlerFunc) {
				err := bootstrap.Initialize(digCon)
				require.NoError(t, err)

				{ // bad password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					req.SetBasicAuth("sigma", "bad_password")
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusUnauthorized, rec.Code)
				}

				{ // correct password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					req.SetBasicAuth("sigma", "sigma")
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // use bearer auth
					ctx := context.Background()
					var userRepository repouser.UserRepository
					require.NoError(t, digCon.Invoke(func(f repouser.UserRepository) { userRepository = f }))
					userObj := &models.User{ID: uuid.NewV7String(), Username: "new-user", Password: new("test"), Email: new("test@gmail.com")}
					err = userRepository.Create(ctx, userObj)
					require.NoError(t, err)

					tokenSvc := genTokenService(t, digCon)
					tokenStr, err := tokenSvc.New(userObj.ID, time.Hour)
					require.NoError(t, err)

					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+tokenStr)
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code)
				}
			},
		},
		{
			name: "skip_auth_check",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() *config.Configuration {
					return &config.Configuration{
						Auth: config.ConfigurationAuth{
							Admin: config.ConfigurationAuthAdmin{
								Username: "sigma",
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
							Jwt: config.ConfigurationAuthJwt{
								PrivateKey: privateKeyString,
							},
						},
						Locker: config.ConfigurationLocker{
							Type:   enums.LockerTypeInmemory,
							Prefix: "sigma-locker",
						},
						Redis: config.ConfigurationRedis{},
						Cache: config.ConfigurationCache{
							Type:   enums.CacherTypeInmemory,
							Prefix: "sigma-cache",
							Inmemory: config.ConfigurationCacheInmemory{
								Size: 100,
							},
						},
					}
				})
				require.NoError(t, err)

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				require.NoError(t, testkit.InitializeIntegration(t, digCon))

				return digCon
			},
			genAuthConfig: func(t *testing.T, c *dig.Container) Config {
				return genAuthConfig(t, c, func(c *gin.Context) bool {
					fmt.Println(c.Request.URL.Path == "/skip", c.Request.URL.Path, "/skip")
					return c.Request.URL.Path == "/skip"
				})
			},
			afterCheck: func(t *testing.T, digCon *dig.Container, middleware gin.HandlerFunc) {
				err := bootstrap.Initialize(digCon)
				require.NoError(t, err)

				{ // skip check
					req := httptest.NewRequest(http.MethodPost, "/skip", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/skip", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // correct password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					req.SetBasicAuth("sigma", "sigma")
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // login with anonymous
					req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{}`))
					req.Header.Set("Content-Type", "application/json")
					rec := httptest.NewRecorder()
					router := gin.New()
					router.Use(middleware)
					router.POST("/test", func(c *gin.Context) {
						c.String(http.StatusOK, "OK")
					})
					router.ServeHTTP(rec, req)
					assert.Equal(t, http.StatusOK, rec.Code)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digCon := tt.genDigCon(t)
			middleware := AuthnWithConfig(tt.genAuthConfig(t, digCon))
			if tt.afterCheck != nil {
				tt.afterCheck(t, digCon, middleware)
			}
		})
	}
}
