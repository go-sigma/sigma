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

package healthz

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testOKResource struct{}

func (t testOKResource) HealthCheck() error {
	return nil
}

type testFailedResource struct{}

func (t testFailedResource) HealthCheck() error {
	return fmt.Errorf("failed")
}

func TestHealthzOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var ok testOKResource
	router := gin.New()
	router.Use(Healthz(ok))
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthzSkipResourceChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var ok testFailedResource
	router := gin.New()
	router.Use(Healthz(ok))
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestReadyzFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var ok testFailedResource
	router := gin.New()
	router.Use(Healthz(ok))
	router.GET("/readyz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHealthzNext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var ok testOKResource
	router := gin.New()
	router.Use(Healthz(ok))
	router.GET("/healthz-test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz-test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
