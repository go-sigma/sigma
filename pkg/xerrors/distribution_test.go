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

package xerrors

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewDSError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.SetBasicAuth("sigma", "sigma1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := NewDSError(c, DSErrCodeBlobUnknown)
	assert.NoError(t, err)
}

func TestDSErrCode(t *testing.T) {
	e := ErrCode{Code: "title", Title: "message"}
	assert.Equal(t, "title: message", e.Error())
}

func TestGenDSErrCodeResourceSizeQuotaExceedNamespace(t *testing.T) {
	assert.Equal(t, "requested access to the size quota is exceed, namespace(library) size quota 100 B/100 B(100%), increasing size is 10 B", GenDSErrCodeResourceSizeQuotaExceedNamespace("library", 100, 100, 10).Title)
}

func TestGenDSErrCodeResourceSizeQuotaExceedRepository(t *testing.T) {
	assert.Equal(t, "requested access to the size quota is exceed, repository(library/alpine) size quota 100 B/100 B(100%), increasing size is 10 B", GenDSErrCodeResourceSizeQuotaExceedRepository("library/alpine", 100, 100, 10).Title)
}

func TestGenDSErrCodeResourceCountQuotaExceedRepository(t *testing.T) {
	assert.Equal(t, "requested access to the resource count quota is exceed, repository(library/alpine) tag count quota is 10", GenDSErrCodeResourceCountQuotaExceedRepository("library/alpine", 10).Title)
}

func TestGenDSErrCodeResourceCountQuotaExceedNamespaceRepository(t *testing.T) {
	assert.Equal(t, "requested access to the resource count quota is exceed, namespace(library) repository count quota is 10", GenDSErrCodeResourceCountQuotaExceedNamespaceRepository("library", 10).Title)
}

func TestGenDSErrCodeResourceCountQuotaExceedNamespaceTag(t *testing.T) {
	assert.Equal(t, "requested access to the resource count quota is exceed, namespace(library) tag count quota is 10", GenDSErrCodeResourceCountQuotaExceedNamespaceTag("library", 10).Title)
}

func TestGenDSErrCodeResourceNotFound(t *testing.T) {
	assert.Equal(t, "Not found", GenDSErrCodeResourceNotFound(errors.New("Not found")).Title)
}

func TestRound(t *testing.T) {
	assert.Equal(t, 1, round(1.4))
	assert.Equal(t, 2, round(1.6))
}

func TestToFixed(t *testing.T) {
	assert.Equal(t, 1.5, toFixed(1.51, 1))
	assert.Equal(t, 1.51, toFixed(1.511, 2))
}
