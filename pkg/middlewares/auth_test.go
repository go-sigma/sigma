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

package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/utils/token"
	"github.com/ximager/ximager/pkg/validators"
)

const (
	privateKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	publicKeyString  = "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHZk1BMEdDU3FHU0liM0RRRUJBUVVBQTRHTkFEQ0JpUUtCZ1FDdm5sMnlNYUZkdDUyRThYSDdrRXZFSG5wbQp6WlZsUFM5YWtlMnlOZCs2bXdVcGVpVDl1alZGcE5idkZBZ2tNNk1Hd3ZZeTdYZFdRcDUwWjl1VUtHdVJSRGUrCnhBdC9uSU5tVkJxUnBTdWdjOE85V0w3NFNveFJXSUoxVXFidzZ2L2hVNytXUjBZTkVNbm1ZaHcyQzV2T0N3N1IKUitCckRPaHloS24rMncxZENRSURBUUFCCi0tLS0tRU5EIFBVQkxJQyBLRVktLS0tLQo="
)

func Test_genWwwAuthenticate(t *testing.T) {
	viper.SetDefault("auth.token.realm", "http://localhost:8080/user/token")
	viper.SetDefault("auth.token.service", "XImager-dev")

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
			want: "Bearer realm=\"http://localhost:8080/user/token\",service=\"XImager-dev\"",
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
	logger.SetLevel(0)
	e := echo.New()
	validators.Initialize(e)
	err := tests.Initialize()
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.jwt.privateKey", privateKeyString)
	viper.SetDefault("auth.jwt.publicKey", publicKeyString)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	hDS := AuthWithConfig(AuthConfig{})(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("ximager", "ximager1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err = hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	req.SetBasicAuth("ximager", "ximager")
	rec1 := httptest.NewRecorder()
	c = e.NewContext(req, rec1)
	err = hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	tokenService, err := token.NewTokenService(viper.GetString("auth.jwt.privateKey"), viper.GetString("auth.jwt.publicKey"))
	assert.NoError(t, err)

	token, err := tokenService.New(&models.User{
		Username: "ximager",
	}, time.Hour)
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec2 := httptest.NewRecorder()
	c = e.NewContext(req, rec2)
	err = hDS(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
}
