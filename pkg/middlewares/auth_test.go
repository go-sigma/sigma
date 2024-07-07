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

package middlewares

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
	"github.com/go-sigma/sigma/pkg/validators"
)

const (
	privateKeyString        = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	privateKeyStringInvalid = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo"
)

func Test_genWwwAuthenticate(t *testing.T) {
	logger.SetLevel("debug")

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
			Token: configs.ConfigurationAuthToken{
				Realm:   "http://localhost:8080/user/token",
				Service: "sigma-dev",
			},
		},
	}
	configs.SetConfiguration(config)

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
			if got := genWwwAuthenticate(tt.args.host, tt.args.schema); got != tt.want {
				t.Errorf("genWwwAuthenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthWithConfig(t *testing.T) {
	logger.SetLevel("debug")

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

	assert.NoError(t, inits.Initialize(configs.Configuration{
		Auth: configs.ConfigurationAuth{
			Admin: configs.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@gmail.com",
			},
		},
	}))

	hDS := AuthWithConfig(AuthConfig{})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	req.SetBasicAuth("sigma", "sigma")
	rec1 := httptest.NewRecorder()
	c = e.NewContext(req, rec1)
	err = hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	tokenService, err := token.NewTokenService(privateKeyString)
	assert.NoError(t, err)

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	ctx := context.Background()
	userObj := &models.User{Username: "post-namespace", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	err = userService.Create(ctx, userObj)
	assert.NoError(t, err)

	token, err := tokenService.New(userObj.ID, time.Hour)
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec2 := httptest.NewRecorder()
	c = e.NewContext(req, rec2)
	err = hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestAuthWithConfigSkipper(t *testing.T) {
	var config = AuthConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/skip"
		},
	}
	mr := AuthWithConfig(config)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/skip", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthWithConfigPrivateKey(t *testing.T) {
	logger.SetLevel("debug")

	config := &configs.Configuration{
		Auth: configs.ConfigurationAuth{
			Admin: configs.ConfigurationAuthAdmin{
				Username: "sigma",
				Password: "sigma",
				Email:    "sigma@gmail.com",
			},
			Jwt: configs.ConfigurationAuthJwt{
				PrivateKey: privateKeyStringInvalid,
			},
			Token: configs.ConfigurationAuthToken{
				Realm:   "http://localhost:8080/user/token",
				Service: "sigma-dev",
			},
		},
	}
	configs.SetConfiguration(config)

	mr := AuthWithConfig(AuthConfig{})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mr1 := AuthWithConfig(AuthConfig{DS: true})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	err = mr1(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAuthWithConfigInvalidBasicAuth(t *testing.T) {
	logger.SetLevel("debug")

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
			Token: configs.ConfigurationAuthToken{
				Realm:   "http://localhost:8080/user/token",
				Service: "sigma-dev",
			},
		},
	}
	configs.SetConfiguration(config)

	mr := AuthWithConfig(AuthConfig{})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Basic dGVzdA===")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	mr1 := AuthWithConfig(AuthConfig{DS: true})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	err = mr1(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
