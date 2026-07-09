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
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/testkit"
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

func TestToken(t *testing.T) {
	logger.SetLevel("debug")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	digCon := dig.New()
	cfg := config.Configuration{
		Auth: config.ConfigurationAuth{
			Jwt: config.ConfigurationAuthJwt{
				PrivateKey: privateKeyString,
			},
		},
	}
	err := digCon.Provide(func() *config.Configuration {
		return &cfg
	})
	require.NoError(t, err)

	const tokenStr = "mock-token-string" // nolint: gosec
	tokenSvc := token.NewMockService(ctrl)
	tokenSvc.EXPECT().New(gomock.Any(), gomock.Any()).DoAndReturn(func(id string, expire time.Duration) (string, error) {
		return tokenStr, nil
	})
	err = digCon.Provide(func() token.Service {
		return tokenSvc
	})
	require.NoError(t, err)

	e := testkit.NewGin()

	require.NoError(t, digCon.Provide(func() *gin.Engine { return e }))

	handler := &handler{
		Config:   &cfg,
		TokenSvc: tokenSvc,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("sigma", "sigma")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set(consts.ContextUser, &models.User{ID: "1", Username: "test"})
	handler.Token(c)
	require.NotNil(t, rec.Body)
	require.Equal(t, tokenStr, gjson.Get(rec.Body.String(), "token").String())
}
