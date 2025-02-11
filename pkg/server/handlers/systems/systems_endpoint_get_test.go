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
)

func TestGetEndpoint(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *configs.Configuration {
		return &configs.Configuration{
			HTTP: configs.ConfigurationHTTP{
				Endpoint: "http://127.0.0.1:3001",
			},
		}
	}))

	e := tests.NewEcho()
	require.NoError(t, digCon.Provide(func() *echo.Echo { return e }))

	handler := handlerNew(digCon)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := handler.GetEndpoint(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)
	response := rec.Body.Bytes()
	assert.Equal(t, "http://127.0.0.1:3001", gjson.GetBytes(response, "endpoint").String())
}
