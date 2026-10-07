// Copyright 2026 sigma
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

package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/server/validators"
)

type bindRequest struct {
	Source string `json:"source" query:"source" param:"source" validate:"required"`
	Body   string `json:"body" validate:"required"`
	Page   *int   `json:"page" query:"page"`
}

type sourceRequest struct {
	Source string `param:"source" validate:"required"`
}

type emptyRequest struct{}

type badQueryRequest struct {
	Bad func() `query:"bad"`
}

type badParamRequest struct {
	Bad func() `param:"bad"`
}

type textRequest struct {
	Body string `json:"body"`
}

func TestBindRequestBindsBodyQueryAndPathBeforeValidation(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.POST("/items/:source", func(c *gin.Context) {
		var req bindRequest
		require.NoError(t, server.BindRequest(c, &req))
		require.Equal(t, "path", req.Source)
		require.Equal(t, "payload", req.Body)
		require.NotNil(t, req.Page)
		require.Equal(t, 2, *req.Page)
		c.Status(http.StatusNoContent)
	})

	body := bytes.NewBufferString(`{"source":"body","body":"payload"}`)
	req := httptest.NewRequest(http.MethodPost, "/items/path?source=query&page=2", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestBindRequestReturnsMalformedJSONError(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.POST("/items/:source", func(c *gin.Context) {
		var req bindRequest
		require.Error(t, server.BindRequest(c, &req))
		c.Status(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodPost, "/items/path", bytes.NewBufferString(`{"body":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWrapRequestReturnsBadRequestForValidationError(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.GET("/items", server.WrapRequest(func(c *gin.Context, req *sourceRequest) {
		t.Fatalf("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), errcode.HTTPErrCodeBadRequest.Code)
}

func TestWrapRequestKeepsHandlerOwnedErrorResponse(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.GET("/items", server.WrapRequest(func(c *gin.Context, req *emptyRequest) {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), errcode.HTTPErrCodeNotFound.Code)
}

func TestBindRequestReturnsQueryMappingError(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.GET("/items", func(c *gin.Context) {
		var req badQueryRequest
		require.Error(t, server.BindRequest(c, &req))
		c.Status(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodGet, "/items?bad=value", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBindRequestReturnsParamMappingError(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.GET("/items/:bad", func(c *gin.Context) {
		var req badParamRequest
		require.Error(t, server.BindRequest(c, &req))
		c.Status(http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/value", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBindRequestSkipsValidationWhenValidatorNil(t *testing.T) {
	setupValidator(t)

	original := binding.Validator
	binding.Validator = nil
	defer func() { binding.Validator = original }()

	router := gin.New()
	router.GET("/items", func(c *gin.Context) {
		var req sourceRequest
		require.NoError(t, server.BindRequest(c, &req))
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestBindRequestIgnoresNonJSONBody(t *testing.T) {
	setupValidator(t)

	router := gin.New()
	router.POST("/items", func(c *gin.Context) {
		var req textRequest
		require.NoError(t, server.BindRequest(c, &req))
		require.Empty(t, req.Body)
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBufferString("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func setupValidator(t *testing.T) {
	t.Helper()
	require.NoError(t, validators.Initialize())
}
