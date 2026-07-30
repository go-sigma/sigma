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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/service/systems"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestGetConfig(t *testing.T) {
	digCon := dig.New()
	cfg := config.Configuration{
		Auth: config.ConfigurationAuth{
			Anonymous: config.ConfigurationAuthAnonymous{
				Enabled: true,
			},
		},
	}
	require.NoError(t, digCon.Provide(func() *config.Configuration {
		return &cfg
	}))

	e := testkit.NewGin()
	require.NoError(t, digCon.Provide(func() *gin.Engine { return e }))

	require.NoError(t, systems.NewService(digCon))
	var systemsSvc systems.Service
	require.NoError(t, digCon.Invoke(func(service systems.Service) {
		systemsSvc = service
	}))
	handler := &handler{SystemsSvc: systemsSvc}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	handler.GetConfig(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	response := rec.Body.Bytes()
	assert.Equal(t, true, gjson.GetBytes(response, "anonymous").Bool())
}
