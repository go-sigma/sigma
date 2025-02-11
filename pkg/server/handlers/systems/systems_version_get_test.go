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

package systems

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
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/version"
)

func TestGetVersion(t *testing.T) {
	// e := echo.New()
	// validators.Initialize(e)
	// assert.NoError(t, tests.Initialize(t))
	// assert.NoError(t, tests.DB.Init())
	// defer func() {
	// 	conn, err := dal.DB.DB()
	// 	assert.NoError(t, err)
	// 	assert.NoError(t, conn.Close())
	// 	assert.NoError(t, tests.DB.DeInit())
	// }()

	// ctrl := gomock.NewController(t)
	// defer ctrl.Finish()

	// systemHandler := handlerNew(inject{})

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *configs.Configuration {
		return &configs.Configuration{
			Auth: configs.ConfigurationAuth{
				Anonymous: configs.ConfigurationAuthAnonymous{
					Enabled: true,
				},
			},
		}
	}))

	e := tests.NewEcho()
	require.NoError(t, digCon.Provide(func() *echo.Echo { return e }))

	handler := handlerNew(digCon)

	ver := "1.0.0"
	version.Version = ver

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler.GetVersion(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	response := rec.Body.Bytes()
	assert.Equal(t, ver, gjson.GetBytes(response, "version").String())
}
