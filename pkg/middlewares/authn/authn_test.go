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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

const (
	privateKeyString = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUNkZ0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQW1Bd2dnSmNBZ0VBQW9HQkFNYmFUZTlsSitBZnQwcWUKMUc5UTF2NWhZLzhFendWZGtXY3FETXlQN2YxVzBwTmxQT1JiSUhqbnp1LytldDlRM1NXcTlWNmF3QkN6M28vSgpZaUF5KzBMYmQ1NmJVcU9aWUxsSTlER1hBd2xrZGR4RWdhNG1CeW8zMmhsblRsak5Kdnc2ckJuYlMrbDIvQzl2CnRWSllpZXlVbExEa2crMExjQTVBZWtYcXlFTTNBZ01CQUFFQ2dZQjJBWjgyYmlWWHovcUtBZSszajVYR3FDMGIKYmRNZE1BWFYzeEp4WXdpc3l4VnoreVJEc0FCNVA3ZUNuTloyS0JyVSs3dFpSU1N0eE5CVExBTmJjR1hDbmpXdQo2ZzFnd0k1bXRQOGFRVzdmUmRETkVlV1NLN0Z6M3BlV2F4UnFzcGpRRXZlcXo5dFZsVnBUbE1ZaDNBcnJWUU9uCm9rWWgzSVJLZitRS0g3MkNJUUpCQVBrMUt6cVNnTWZlRXBtUEEvM3NRb2Q5YmQwTlFpeG05SUFjYUwvWTV1YzYKSEJ3R1pMSmhUdTN1Tmd1aEsvUWJ4NlFvbXNSN2E5cS94WEMydGprMUZtMENRUURNUmNtQXZSMnlvYksvZUhURApQQnR0clFnTFZtRXIrRStGS25ubDNlOHQ5eTNScVh6RCsyUnNrUjFJb2QwR1JYNzFwRi9PNXFUdGo1Mi9yN0liCmhzbXpBa0VBZytvTUZ2WWo2eWgzU2dlMVFqMUV2am03NVE0Mm9CQmpqa2o3ZmNvUCtBZi9oeW92TldsakFYbGQKN0d3Rk96TlZTMlVlLzdDaFYrcTVWYit4MTdodFJRSkFRcmVrYWF6YTcwWUsyS2lpRWtZbWV6cmhmcnAyd0dLNApyampDV1lhVUlRSXpiK0FZaFBZdHhadmI0YVlrUjNFWlYyZVpkejB6cnZlU1FWSkVMT05vS3dKQUJUWElEVStWCjVvcDdIb2FRSlBXOUYyYkY5cE9kTGlBdzFMeC9kTGJ5TnNhaHZLNEkxZDdtZVZIcFhENDZ0ekF0ZW1Gck1qbisKRlNkQkl5YWNNYndza3c9PQotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
)

func TestGenWwwAuthenticate(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()
	err := digCon.Provide(func() configs.Configuration {
		return configs.Configuration{
			Auth: configs.ConfigurationAuth{
				Admin: configs.ConfigurationAuthAdmin{
					Username: "sigma",
					Password: "sigma",
					Email:    "sigma@tosone.cn",
				},
				Jwt: configs.ConfigurationAuthJwt{
					PrivateKey: privateKeyString,
				},
				Token: configs.ConfigurationAuthToken{
					Realm:   "http://localhost:8080/user/token",
					Service: "sigma-dev",
				},
			},
		}
	})
	require.NoError(t, err)

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
			if got := genWwwAuthenticate(digCon, tt.args.host, tt.args.schema); got != tt.want {
				t.Errorf("genWwwAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthWithConfig(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name          string
		genDigCon     func(*testing.T) *dig.Container
		genAuthConfig func(*testing.T, *dig.Container) Config
		afterCheck    func(*testing.T, *dig.Container, echo.MiddlewareFunc)
		afterEach     func(*testing.T, *dig.Container)
	}{
		{
			name: "normal",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
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
						Locker: configs.ConfigurationLocker{
							Type:   enums.LockerTypeRedis,
							Prefix: "sigma-locker",
							Redis:  configs.ConfigurationLockerRedis{},
						},
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis://:sigma@localhost:6379/0",
						},
						Cache: configs.ConfigurationCache{
							Type:   enums.CacherTypeRedis,
							Prefix: "sigma-cache",
							Redis: configs.ConfigurationCacheRedis{
								Ttl: time.Hour * 24 * 7,
							},
						},
						Database: configs.ConfigurationDatabase{
							Type: enums.DatabaseSqlite3,
							Sqlite3: configs.ConfigurationDatabaseSqlite3{
								Path: fmt.Sprintf("%s.db", strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")),
							},
							Mysql: configs.ConfigurationDatabaseMysql{
								Host:     "127.0.0.1",
								Port:     3306,
								Username: "root",
								Password: "sigma",
								Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
							},
							Postgresql: configs.ConfigurationDatabasePostgresql{
								Host:     "127.0.0.1",
								Port:     5432,
								Username: "sigma",
								Password: "sigma",
								Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
								SslMode:  "disable",
							},
						},
					}
				})
				require.NoError(t, err)

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				testInstance, err := tests.Initialize(t, digCon)
				require.NoError(t, err)

				err = digCon.Provide(func() *tests.Instance {
					return testInstance
				})
				require.NoError(t, err)

				return digCon
			},
			genAuthConfig: func(t *testing.T, c *dig.Container) Config {
				return Config{
					DigCon: c,
				}
			},
			afterCheck: func(t *testing.T, digCon *dig.Container, middleware echo.MiddlewareFunc) {
				handlerFunc := middleware(func(c echo.Context) error {
					return c.String(http.StatusOK, "OK")
				})

				err := inits.Initialize(digCon)
				require.NoError(t, err)

				{ // bad password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					req.SetBasicAuth("sigma", "bad_password")
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err := handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusUnauthorized, rec.Code)
				}

				{ // correct password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					req.SetBasicAuth("sigma", "sigma")
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err := handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // use bearer auth
					ctx := log.Logger.WithContext(context.Background())
					userService := utils.MustGetObjFromDigCon[dao.UserServiceFactory](digCon).New()
					userObj := &models.User{Username: "new-user", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
					err = userService.Create(ctx, userObj)
					require.NoError(t, err)

					tokenService, err := token.New(digCon)
					require.NoError(t, err)
					tokenStr, err := tokenService.New(userObj.ID, time.Hour)
					require.NoError(t, err)

					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					req.Header.Set(echo.HeaderAuthorization, "Bearer "+tokenStr)
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err = handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusOK, rec.Code)
				}
			},
			afterEach: func(t *testing.T, c *dig.Container) {
				require.NoError(t, dal.DeInitialize())
				require.NoError(t, utils.MustGetObjFromDigCon[*tests.Instance](c).DeInitialize())
			},
		},
		{
			name: "skip_auth_check",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
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
						Locker: configs.ConfigurationLocker{
							Type:   enums.LockerTypeRedis,
							Prefix: "sigma-locker",
							Redis:  configs.ConfigurationLockerRedis{},
						},
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							URL:  "redis://:sigma@localhost:6379/0",
						},
						Cache: configs.ConfigurationCache{
							Type:   enums.CacherTypeRedis,
							Prefix: "sigma-cache",
							Redis: configs.ConfigurationCacheRedis{
								Ttl: time.Hour * 24 * 7,
							},
						},
						Database: configs.ConfigurationDatabase{
							Type: enums.DatabaseSqlite3,
							Sqlite3: configs.ConfigurationDatabaseSqlite3{
								Path: fmt.Sprintf("%s.db", strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")),
							},
							Mysql: configs.ConfigurationDatabaseMysql{
								Host:     "127.0.0.1",
								Port:     3306,
								Username: "root",
								Password: "sigma",
								Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
							},
							Postgresql: configs.ConfigurationDatabasePostgresql{
								Host:     "127.0.0.1",
								Port:     5432,
								Username: "sigma",
								Password: "sigma",
								Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
								SslMode:  "disable",
							},
						},
					}
				})
				require.NoError(t, err)

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				testInstance, err := tests.Initialize(t, digCon)
				require.NoError(t, err)

				err = digCon.Provide(func() *tests.Instance {
					return testInstance
				})
				require.NoError(t, err)

				return digCon
			},
			genAuthConfig: func(t *testing.T, c *dig.Container) Config {
				return Config{
					DigCon: c,
					Skipper: func(c echo.Context) bool {
						fmt.Println(c.Request().URL.Path == "/skip", c.Request().URL.Path, "/skip")
						return c.Request().URL.Path == "/skip"
					},
				}
			},
			afterCheck: func(t *testing.T, digCon *dig.Container, middleware echo.MiddlewareFunc) {
				handlerFunc := middleware(func(c echo.Context) error {
					return c.String(http.StatusOK, "OK")
				})

				err := inits.Initialize(digCon)
				require.NoError(t, err)

				{ // skip check
					req := httptest.NewRequest(http.MethodPost, "/skip", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err := handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // correct password
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					req.SetBasicAuth("sigma", "sigma")
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err := handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusOK, rec.Code)
				}

				{ // login with anonymous
					req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{}`))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
					rec := httptest.NewRecorder()
					echo := tests.NewEcho()
					c := echo.NewContext(req, rec)
					err := handlerFunc(c)
					assert.NoError(t, err)
					assert.Equal(t, http.StatusOK, rec.Code)
				}
			},
			afterEach: func(t *testing.T, c *dig.Container) {
				require.NoError(t, dal.DeInitialize())
				require.NoError(t, utils.MustGetObjFromDigCon[*tests.Instance](c).DeInitialize())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digCon := tt.genDigCon(t)
			defer tt.afterEach(t, digCon)
			middleware := AuthnWithConfig(tt.genAuthConfig(t, digCon))
			if tt.afterCheck != nil {
				tt.afterCheck(t, digCon, middleware)
			}
		})
	}
}
