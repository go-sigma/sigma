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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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
	var ok testOKResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthzFailed(t *testing.T) {
	var ok testFailedResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHealthzNext(t *testing.T) {
	var ok testOKResource
	mr := Healthz(ok)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz-test", nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	err := mr(c)
	assert.NoError(t, err)
}
